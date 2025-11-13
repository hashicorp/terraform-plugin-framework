// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

func (s *Server) StateByteChunkResourceType(ctx context.Context, typeName string) (statestore.StateByteChunk, diag.Diagnostics) {
	statebytechunkResourceFuncs, diags := s.StateByteChunkResourceFuncs(ctx)
	statebytechunkResourceFunc, ok := statebytechunkResourceFuncs[typeName]

	if !ok {
		diags.AddError(
			"StateByteChunk Resource Type Not Found",
			fmt.Sprintf("No statebytechunk resource type named %q was found in the provider.", typeName),
		)

		return nil, diags
	}

	return statebytechunkResourceFunc(), nil
}

// StateByteChunkResourceFuncs returns a map of StateByteChunkResource functions. The results are
// cached on first use.
func (s *Server) StateByteChunkResourceFuncs(ctx context.Context) (map[string]func() statestore.StateByteChunk, diag.Diagnostics) {
	provider, ok := s.Provider.(provider.ProviderWithStateByteChunkResources)
	if !ok {
		return nil, nil
	}

	logging.FrameworkTrace(ctx, "Checking StateByteChunkResourceFuncs lock")
	s.statebytechunkResourceFuncsMutex.Lock()
	defer s.statebytechunkResourceFuncsMutex.Unlock()

	if s.statebytechunkResourceFuncs != nil {
		return s.statebytechunkResourceFuncs, s.statebytechunkResourceFuncsDiags
	}

	providerTypeName := s.ProviderTypeName(ctx)
	s.statebytechunkResourceFuncs = make(map[string]func() statestore.StateByteChunkResource)

	logging.FrameworkTrace(ctx, "Calling provider defined StateByteChunkResources")
	statebytechunkResourceFuncSlice := provider.StateByteChunkResources(ctx)
	logging.FrameworkTrace(ctx, "Called provider defined StateByteChunkResources")

	for _, statebytechunkResourceFunc := range statebytechunkResourceFuncSlice {
		statebytechunkResource := statebytechunkResourceFunc()

		metadataReq := resource.MetadataRequest{
			ProviderTypeName: providerTypeName,
		}
		metadataResp := resource.MetadataResponse{}
		statebytechunkResource.Metadata(ctx, metadataReq, &metadataResp)

		typeName := metadataResp.TypeName
		if typeName == "" {
			s.statebytechunkResourceFuncsDiags.AddError(
				"StateByteChunkResource Type Name Missing",
				fmt.Sprintf("The %T StateByteChunkResource returned an empty string from the Metadata method. ", statebytechunkResource)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		logging.FrameworkTrace(ctx, "Found resource type", map[string]interface{}{logging.KeyStateByteChunkResourceType: typeName}) // TODO: y?

		if _, ok := s.statebytechunkResourceFuncs[typeName]; ok {
			s.statebytechunkResourceFuncsDiags.AddError(
				"Duplicate StateByteChunkResource Type Defined",
				fmt.Sprintf("The %s StateByteChunkResource type name was returned for multiple statebytechunk resources. ", typeName)+
					"StateByteChunkResource type names must be unique. "+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		rawV5SchemasResp := statestore.RawV5SchemaResponse{}
		if statebytechunkResourceWithSchemas, ok := statebytechunkResource.(statestore.StateByteChunkResourceWithRawV5Schemas); ok {
			statebytechunkResourceWithSchemas.RawV5Schemas(ctx, statestore.RawV5SchemaRequest{}, &rawV5SchemasResp)
		}

		rawV6SchemasResp := statestore.RawV6SchemaResponse{}
		if statebytechunkResourceWithSchemas, ok := statebytechunkResource.(statestore.StateByteChunkResourceWithRawV6Schemas); ok {
			statebytechunkResourceWithSchemas.RawV6Schemas(ctx, statestore.RawV6SchemaRequest{}, &rawV6SchemasResp)
		}

		resourceFuncs, _ := s.ResourceFuncs(ctx)
		if _, ok := resourceFuncs[typeName]; !ok {
			if (rawV5SchemasResp.ProtoV5Schema == nil || rawV5SchemasResp.ProtoV5IdentitySchema == nil) && (rawV6SchemasResp.ProtoV6Schema == nil || rawV6SchemasResp.ProtoV6IdentitySchema == nil) {
				s.statebytechunkResourceFuncsDiags.AddError(
					"StateByteChunkResource Type Defined without a Matching Managed Resource Type",
					fmt.Sprintf("The %s StateByteChunkResource type name was returned, but no matching managed Resource type was defined. ", typeName)+
						"If the matching managed Resource type is not a framework resource either ProtoV5Schema and ProtoV5IdentitySchema must be specified in the RawV5Schemas method, "+
						"or ProtoV6Schema and ProtoV6IdentitySchema must be specified in the RawV6Schemas method. "+
						"This is always an issue with the provider and should be reported to the provider developers.",
				)
				continue
			}
		}

		s.statebytechunkResourceFuncs[typeName] = statebytechunkResourceFunc
	}

	return s.statebytechunkResourceFuncs, s.statebytechunkResourceFuncsDiags
}

// StateByteChunkResourceMetadatas returns a slice of StateByteChunkResourceMetadata for the GetMetadata
// RPC.
func (s *Server) StateByteChunkResourceMetadatas(ctx context.Context) ([]StateByteChunkResourceMetadata, diag.Diagnostics) {
	statebytechunkResourceFuncs, diags := s.StateByteChunkResourceFuncs(ctx)

	statebytechunkResourceMetadatas := make([]StateByteChunkResourceMetadata, 0, len(statebytechunkResourceFuncs))

	for typeName := range statebytechunkResourceFuncs {
		statebytechunkResourceMetadatas = append(statebytechunkResourceMetadatas, StateByteChunkResourceMetadata{
			TypeName: typeName,
		})
	}

	return statebytechunkResourceMetadatas, diags
}

// StateByteChunkResourceSchema returns the StateByteChunkResource Schema for the given type name and
// caches the result for later StateByteChunkResource operations.
func (s *Server) StateByteChunkResourceSchema(ctx context.Context, typeName string) (fwschema.Schema, diag.Diagnostics) {
	s.statebytechunkResourceSchemasMutex.RLock()
	statebytechunkResourceSchema, ok := s.statebytechunkResourceSchemas[typeName]
	s.statebytechunkResourceSchemasMutex.RUnlock()

	if ok {
		return statebytechunkResourceSchema, nil
	}

	var diags diag.Diagnostics

	statebytechunkResource, statebytechunkResourceDiags := s.StateByteChunkResourceType(ctx, typeName)
	diags.Append(statebytechunkResourceDiags...)
	if diags.HasError() {
		return nil, diags
	}

	schemaReq := statestore.StateByteChunkResourceSchemaRequest{}
	schemaResp := statestore.StateByteChunkResourceSchemaResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined StateByteChunkResourceConfigSchema method", map[string]interface{}{logging.KeyStateByteChunkResourceType: typeName})
	statebytechunkResource.StateByteChunkResourceConfigSchema(ctx, schemaReq, &schemaResp)
	logging.FrameworkTrace(ctx, "Called provider defined StateByteChunkResourceConfigSchema method", map[string]interface{}{logging.KeyStateByteChunkResourceType: typeName})

	diags.Append(schemaResp.Diagnostics...)
	if diags.HasError() {
		return schemaResp.Schema, diags
	}

	s.statebytechunkResourceSchemasMutex.Lock()

	if s.statebytechunkResourceSchemas == nil {
		s.statebytechunkResourceSchemas = make(map[string]fwschema.Schema)
	}

	s.statebytechunkResourceSchemas[typeName] = schemaResp.Schema

	s.statebytechunkResourceSchemasMutex.Unlock()

	return schemaResp.Schema, diags
}

// StateByteChunkResourceSchemas returns a map of StateByteChunkResource Schemas for the
// GetProviderSchema RPC without caching since not all schemas are guaranteed to
// be necessary for later provider operations. The schema implementations are
// also validated.
func (s *Server) StateByteChunkResourceSchemas(ctx context.Context) (map[string]fwschema.Schema, diag.Diagnostics) {
	statebytechunkResourceSchemas := make(map[string]fwschema.Schema)
	statebytechunkResourceFuncs, diags := s.StateByteChunkResourceFuncs(ctx)

	for typeName, statebytechunkResourceFunc := range statebytechunkResourceFuncs {
		statebytechunkResource := statebytechunkResourceFunc()
		schemaReq := statestore.StateByteChunkResourceSchemaRequest{}
		schemaResp := statestore.StateByteChunkResourceSchemaResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined StateByteChunkResource Schemas", map[string]interface{}{logging.KeyStateByteChunkResourceType: typeName})
		statebytechunkResource.StateByteChunkResourceConfigSchema(ctx, schemaReq, &schemaResp)
		logging.FrameworkTrace(ctx, "Called provider defined StateByteChunkResource Schemas", map[string]interface{}{logging.KeyStateByteChunkResourceType: typeName})

		diags.Append(schemaResp.Diagnostics...)
		if schemaResp.Diagnostics.HasError() {
			continue
		}

		validateDiags := schemaResp.Schema.ValidateImplementation(ctx)
		diags.Append(validateDiags...)
		if validateDiags.HasError() {
			continue
		}

		statebytechunkResourceSchemas[typeName] = schemaResp.Schema
	}

	return statebytechunkResourceSchemas, diags
}
