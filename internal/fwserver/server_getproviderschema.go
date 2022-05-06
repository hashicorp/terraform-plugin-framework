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
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetSchema")
	providerSchema, diags := s.Provider.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetSchema")

	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	resp.Provider = &providerSchema

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

	// TODO: Cache GetDataSources call
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/299
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetResources")
	resourceSchemas, diags := s.Provider.GetResources(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetResources")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(resourceSchemas) > 0 {
		resp.ResourceSchemas = map[string]*tfsdk.Schema{}
	}

	for k, v := range resourceSchemas {
		// KeyResourceType field only necessary here since we are in GetProviderSchema RPC
		logging.FrameworkTrace(ctx, "Found resource type", map[string]interface{}{logging.KeyResourceType: k})

		logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema", map[string]interface{}{logging.KeyResourceType: k})
		schema, diags := v.GetSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema", map[string]interface{}{logging.KeyResourceType: k})

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.ResourceSchemas[k] = &schema
	}

	// TODO: Cache GetDataSources call
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/299
	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetDataSources")
	dataSourceSchemas, diags := s.Provider.GetDataSources(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetDataSources")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(dataSourceSchemas) > 0 {
		resp.DataSourceSchemas = map[string]*tfsdk.Schema{}
	}

	for k, v := range dataSourceSchemas {
		// KeyDataSourceType field only necessary here since we are in GetProviderSchema RPC
		logging.FrameworkTrace(ctx, "Found data source type", map[string]interface{}{logging.KeyDataSourceType: k})

		logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType GetSchema", map[string]interface{}{logging.KeyDataSourceType: k})
		schema, diags := v.GetSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined DataSourceType GetSchema", map[string]interface{}{logging.KeyDataSourceType: k})

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.DataSourceSchemas[k] = &schema
	}
}
