// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ListResourceFuncs returns a map of ListResource functions. The results are
// cached on first use.
func (s *Server) ListResourceFuncs(ctx context.Context) (map[string]func() list.ListResource, diag.Diagnostics) {
	provider, ok := s.Provider.(provider.ProviderWithListResources)
	if !ok {
		return nil, nil
	}

	logging.FrameworkTrace(ctx, "Checking ListResourceTypes lock")
	s.listResourceFuncsMutex.Lock()
	defer s.listResourceFuncsMutex.Unlock()

	if s.listResourceFuncs != nil {
		return s.listResourceFuncs, s.resourceTypesDiags
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

		if _, ok := s.resourceFuncs[typeName]; !ok {
			s.listResourceFuncsDiags.AddError(
				"ListResource Type Defined without a Matching Managed Resource Type",
				fmt.Sprintf("The %s ListResource type name was returned, but no matching managed Resource type was defined. ", typeName)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		s.listResourceFuncs[typeName] = listResourceFunc
	}

	return s.listResourceFuncs, s.listResourceFuncsDiags
}

// ListResourceMetadatas returns a slice of ListResourceMetadata for the GetMetadata
// RPC.
func (s *Server) ListResourceMetadatas(ctx context.Context) ([]ListResourceMetadata, diag.Diagnostics) {
	resourceFuncs, diags := s.ListResourceFuncs(ctx)

	resourceMetadatas := make([]ListResourceMetadata, 0, len(resourceFuncs))

	for typeName := range resourceFuncs {
		resourceMetadatas = append(resourceMetadatas, ListResourceMetadata{
			TypeName: typeName,
		})
	}

	return resourceMetadatas, diags
}
