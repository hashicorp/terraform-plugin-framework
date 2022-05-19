package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// GetProviderSchemaRequest is the framework server request for the
// GetProviderSchema RPC.
type GetProviderSchemaRequest struct{}

// GetProviderSchemaResponse is the framework server response for the
// GetProviderSchema RPC.
type GetProviderSchemaResponse struct {
	Provider          *tfsdk.Schema
	ProviderMeta      *tfsdk.Schema
	ResourceSchemas   map[string]*tfsdk.Schema
	DataSourceSchemas map[string]*tfsdk.Schema
	Diagnostics       diag.Diagnostics
}

// GetProviderSchema implements the framework server GetProviderSchema RPC.
func (s *Server) GetProviderSchema(ctx context.Context, req *GetProviderSchemaRequest, resp *GetProviderSchemaResponse) {
	providerSchema, diags := s.ProviderSchema(ctx)

	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	resp.Provider = providerSchema

	if pm, ok := s.Provider.(tfsdk.ProviderWithProviderMeta); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithProviderMeta")

		logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
		providerMetaSchema, diags := pm.GetMetaSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.ProviderMeta = &providerMetaSchema
	}

	resourceSchemas, diags := s.ResourceSchemas(ctx)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.ResourceSchemas = resourceSchemas

	dataSourceSchemas, diags := s.DataSourceSchemas(ctx)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSourceSchemas = dataSourceSchemas
}
