// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetMetadataRequest is the framework server request for the
// GetMetadata RPC.
type GetMetadataRequest struct{}

// GetMetadataResponse is the framework server response for the
// GetMetadata RPC.
type GetMetadataResponse struct {
	DataSources        []DataSourceMetadata
	Diagnostics        diag.Diagnostics
	EphemeralResources []EphemeralResourceMetadata
	Functions          []FunctionMetadata
	ListResources      []ListResourceMetadata
	Resources          []ResourceMetadata
	ServerCapabilities *ServerCapabilities
}

// DataSourceMetadata is the framework server equivalent of the
// tfprotov5.DataSourceMetadata and tfprotov6.DataSourceMetadata types.
type DataSourceMetadata struct {
	// TypeName is the name of the data resource.
	TypeName string
}

// EphemeralResourceMetadata is the framework server equivalent of the
// tfprotov5.EphemeralResourceMetadata and tfprotov6.EphemeralResourceMetadata types.
type EphemeralResourceMetadata struct {
	// TypeName is the name of the ephemeral resource.
	TypeName string
}

// FunctionMetadata is the framework server equivalent of the
// tfprotov5.FunctionMetadata and tfprotov6.FunctionMetadata types.
type FunctionMetadata struct {
	// Name is the name of the function.
	Name string
}

// ResourceMetadata is the framework server equivalent of the
// tfprotov5.ResourceMetadata and tfprotov6.ResourceMetadata types.
type ResourceMetadata struct {
	// TypeName is the name of the managed resource.
	TypeName string
}

// ListResourceMetadata is the framework server equivalent of the
// tfprotov5.ListResourceMetadata and tfprotov6.ListResourceMetadata types.
type ListResourceMetadata struct {
	// TypeName is the name of the list resource.
	TypeName string
}

// GetMetadata implements the framework server GetMetadata RPC.
func (s *Server) GetMetadata(ctx context.Context, req *GetMetadataRequest, resp *GetMetadataResponse) {
	resp.DataSources = []DataSourceMetadata{}
	resp.EphemeralResources = []EphemeralResourceMetadata{}
	resp.Functions = []FunctionMetadata{}
	resp.ListResources = []ListResourceMetadata{}
	resp.Resources = []ResourceMetadata{}

	resp.ServerCapabilities = s.ServerCapabilities()

	datasourceMetadatas, diags := s.DataSourceMetadatas(ctx)
	resp.Diagnostics.Append(diags...)

	ephemeralResourceMetadatas, diags := s.EphemeralResourceMetadatas(ctx)
	resp.Diagnostics.Append(diags...)

	functionMetadatas, diags := s.FunctionMetadatas(ctx)
	resp.Diagnostics.Append(diags...)

	resourceMetadatas, diags := s.ResourceMetadatas(ctx)
	resp.Diagnostics.Append(diags...)

	listResourceMetadatas, diags := s.ListResourceMetadatas(ctx)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSources = datasourceMetadatas
	resp.EphemeralResources = ephemeralResourceMetadatas
	resp.Functions = functionMetadatas
	resp.ListResources = listResourceMetadatas
	resp.Resources = resourceMetadatas
}
