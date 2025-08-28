// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (s *Server) ListResourceType(ctx context.Context, typeName string) (list.ListResource, diag.Diagnostics) {
	listResourceFuncs, diags := s.ListResourceFuncs(ctx)
	listResourceFunc, ok := listResourceFuncs[typeName]

	if !ok {
		diags.AddError(
			"List Resource Type Not Found",
			fmt.Sprintf("No list resource type named %q was found in the provider.", typeName),
		)

		return nil, diags
	}

	return listResourceFunc(), nil
}

// ListResourceFuncs returns a map of ListResource functions. The results are
// cached on first use.
func (s *Server) ListResourceFuncs(ctx context.Context) (map[string]func() list.ListResource, diag.Diagnostics) {
	provider, ok := s.Provider.(provider.ProviderWithListResources)
	if !ok {
		return nil, nil
	}

	logging.FrameworkTrace(ctx, "Checking ListResourceFuncs lock")
	s.listResourceFuncsMutex.Lock()
	defer s.listResourceFuncsMutex.Unlock()

	if s.listResourceFuncs != nil {
		return s.listResourceFuncs, s.listResourceFuncsDiags
	}

	providerTypeName := s.ProviderTypeName(ctx)
	s.listResourceFuncs = make(map[string]func() list.ListResource)

	logging.FrameworkTrace(ctx, "Calling provider defined ListResources")
	listResourceFuncSlice := provider.ListResources(ctx)
	logging.FrameworkTrace(ctx, "Called provider defined ListResources")

	for _, listResourceFunc := range listResourceFuncSlice {
		listResource := listResourceFunc()

		metadataReq := resource.MetadataRequest{
			ProviderTypeName: providerTypeName,
		}
		metadataResp := resource.MetadataResponse{}
		listResource.Metadata(ctx, metadataReq, &metadataResp)

		typeName := metadataResp.TypeName
		if typeName == "" {
			s.listResourceFuncsDiags.AddError(
				"ListResource Type Name Missing",
				fmt.Sprintf("The %T ListResource returned an empty string from the Metadata method. ", listResource)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		logging.FrameworkTrace(ctx, "Found resource type", map[string]interface{}{logging.KeyListResourceType: typeName}) // TODO: y?

		if _, ok := s.listResourceFuncs[typeName]; ok {
			s.listResourceFuncsDiags.AddError(
				"Duplicate ListResource Type Defined",
				fmt.Sprintf("The %s ListResource type name was returned for multiple list resources. ", typeName)+
					"ListResource type names must be unique. "+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		resourceFuncs, _ := s.ResourceFuncs(ctx)
		if _, ok := resourceFuncs[typeName]; !ok {
			if metadataResp.ProtoV5Schema == nil || metadataResp.ProtoV5IdentitySchema == nil {
				s.listResourceFuncsDiags.AddError(
					"ListResource Type Defined without a Matching Managed Resource Type",
					fmt.Sprintf("The %s ListResource type name was returned, but no matching managed Resource type was defined. ", typeName)+
						"If the matching managed Resource type is a legacy resource, ProtoV5Schema and ProtoV5IdentitySchema must be specified in the Metadata method. "+
						"This is always an issue with the provider and should be reported to the provider developers.",
				)
				continue
			}
		}

		s.listResourceFuncs[typeName] = listResourceFunc
	}

	return s.listResourceFuncs, s.listResourceFuncsDiags
}

// ListResourceMetadatas returns a slice of ListResourceMetadata for the GetMetadata
// RPC.
func (s *Server) ListResourceMetadatas(ctx context.Context) ([]ListResourceMetadata, diag.Diagnostics) {
	listResourceFuncs, diags := s.ListResourceFuncs(ctx)

	listResourceMetadatas := make([]ListResourceMetadata, 0, len(listResourceFuncs))

	for typeName := range listResourceFuncs {
		listResourceMetadatas = append(listResourceMetadatas, ListResourceMetadata{
			TypeName: typeName,
		})
	}

	return listResourceMetadatas, diags
}

// ListResourceSchema returns the ListResource Schema for the given type name and
// caches the result for later ListResource operations.
func (s *Server) ListResourceSchema(ctx context.Context, typeName string) (fwschema.Schema, diag.Diagnostics) {
	s.listResourceSchemasMutex.RLock()
	listResourceSchema, ok := s.listResourceSchemas[typeName]
	s.listResourceSchemasMutex.RUnlock()

	if ok {
		return listResourceSchema, nil
	}

	var diags diag.Diagnostics

	listResource, listResourceDiags := s.ListResourceType(ctx, typeName)
	diags.Append(listResourceDiags...)
	if diags.HasError() {
		return nil, diags
	}

	schemaReq := list.ListResourceSchemaRequest{}
	schemaResp := list.ListResourceSchemaResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined ListResourceConfigSchema method", map[string]interface{}{logging.KeyListResourceType: typeName})
	listResource.ListResourceConfigSchema(ctx, schemaReq, &schemaResp)
	logging.FrameworkTrace(ctx, "Called provider defined ListResourceConfigSchema method", map[string]interface{}{logging.KeyListResourceType: typeName})

	diags.Append(schemaResp.Diagnostics...)
	if diags.HasError() {
		return schemaResp.Schema, diags
	}

	s.listResourceSchemasMutex.Lock()

	if s.listResourceSchemas == nil {
		s.listResourceSchemas = make(map[string]fwschema.Schema)
	}

	s.listResourceSchemas[typeName] = schemaResp.Schema

	s.listResourceSchemasMutex.Unlock()

	return schemaResp.Schema, diags
}

// ListResourceSchemas returns a map of ListResource Schemas for the
// GetProviderSchema RPC without caching since not all schemas are guaranteed to
// be necessary for later provider operations. The schema implementations are
// also validated.
func (s *Server) ListResourceSchemas(ctx context.Context) (map[string]fwschema.Schema, diag.Diagnostics) {
	listResourceSchemas := make(map[string]fwschema.Schema)
	listResourceFuncs, diags := s.ListResourceFuncs(ctx)

	for typeName, listResourceFunc := range listResourceFuncs {
		listResource := listResourceFunc()
		schemaReq := list.ListResourceSchemaRequest{}
		schemaResp := list.ListResourceSchemaResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined ListResource Schemas", map[string]interface{}{logging.KeyListResourceType: typeName})
		listResource.ListResourceConfigSchema(ctx, schemaReq, &schemaResp)
		logging.FrameworkTrace(ctx, "Called provider defined ListResource Schemas", map[string]interface{}{logging.KeyListResourceType: typeName})

		diags.Append(schemaResp.Diagnostics...)
		if schemaResp.Diagnostics.HasError() {
			continue
		}

		validateDiags := schemaResp.Schema.ValidateImplementation(ctx)
		diags.Append(validateDiags...)
		if validateDiags.HasError() {
			continue
		}

		listResourceSchemas[typeName] = schemaResp.Schema
	}

	return listResourceSchemas, diags
}
