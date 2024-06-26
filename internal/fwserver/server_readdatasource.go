// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ReadDataSourceRequest is the framework server request for the
// ReadDataSource RPC.
type ReadDataSourceRequest struct {
	ClientCapabilities datasource.ReadClientCapabilities
	Config             *tfsdk.Config
	DataSourceSchema   fwschema.Schema
	DataSource         datasource.DataSource
	ProviderMeta       *tfsdk.Config
}

// ReadDataSourceResponse is the framework server response for the
// ReadDataSource RPC.
type ReadDataSourceResponse struct {
	Deferred    *datasource.Deferred
	Diagnostics diag.Diagnostics
	State       *tfsdk.State
}

// ReadDataSource implements the framework server ReadDataSource RPC.
func (s *Server) ReadDataSource(ctx context.Context, req *ReadDataSourceRequest, resp *ReadDataSourceResponse) {
	if req == nil {
		return
	}

	if s.deferred != nil {
		logging.FrameworkDebug(ctx, "Provider has deferred response configured, automatically returning deferred response.",
			map[string]interface{}{
				logging.KeyDeferredReason: s.deferred.Reason.String(),
			},
		)
		// Send an unknown value for the data source. This will replace any configured values
		// for ease of implementation as Terraform Core currently does not use these values for
		// deferred actions, but this design could change in the future.
		resp.State = &tfsdk.State{
			Raw:    tftypes.NewValue(req.DataSourceSchema.Type().TerraformType(ctx), tftypes.UnknownValue),
			Schema: req.DataSourceSchema,
		}
		resp.Deferred = &datasource.Deferred{
			Reason: datasource.DeferredReason(s.deferred.Reason),
		}
		return
	}

	if dataSourceWithConfigure, ok := req.DataSource.(datasource.DataSourceWithConfigure); ok {
		logging.FrameworkTrace(ctx, "DataSource implements DataSourceWithConfigure")

		configureReq := datasource.ConfigureRequest{
			ProviderData: s.DataSourceConfigureData,
		}
		configureResp := datasource.ConfigureResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined DataSource Configure")
		dataSourceWithConfigure.Configure(ctx, configureReq, &configureResp)
		logging.FrameworkTrace(ctx, "Called provider defined DataSource Configure")

		resp.Diagnostics.Append(configureResp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	readReq := datasource.ReadRequest{
		ClientCapabilities: req.ClientCapabilities,
		Config: tfsdk.Config{
			Schema: req.DataSourceSchema,
		},
	}
	readResp := datasource.ReadResponse{
		State: tfsdk.State{
			Schema: req.DataSourceSchema,
		},
	}

	if req.Config != nil {
		readReq.Config = *req.Config
		readResp.State.Raw = req.Config.Raw.Copy()
	}

	if req.ProviderMeta != nil {
		readReq.ProviderMeta = *req.ProviderMeta
	}

	logging.FrameworkTrace(ctx, "Calling provider defined DataSource Read")
	req.DataSource.Read(ctx, readReq, &readResp)
	logging.FrameworkTrace(ctx, "Called provider defined DataSource Read")

	resp.Diagnostics = readResp.Diagnostics
	resp.State = &readResp.State
	resp.Deferred = readResp.Deferred

	if resp.Diagnostics.HasError() {
		return
	}

	semanticEqualityReq := SchemaSemanticEqualityRequest{
		PriorData: fwschemadata.Data{
			Description:    fwschemadata.DataDescriptionConfiguration,
			Schema:         req.Config.Schema,
			TerraformValue: req.Config.Raw.Copy(),
		},
		ProposedNewData: fwschemadata.Data{
			Description:    fwschemadata.DataDescriptionState,
			Schema:         resp.State.Schema,
			TerraformValue: resp.State.Raw.Copy(),
		},
	}
	semanticEqualityResp := &SchemaSemanticEqualityResponse{
		NewData: semanticEqualityReq.ProposedNewData,
	}

	SchemaSemanticEquality(ctx, semanticEqualityReq, semanticEqualityResp)

	resp.Diagnostics.Append(semanticEqualityResp.Diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	if semanticEqualityResp.NewData.TerraformValue.Equal(resp.State.Raw) {
		return
	}

	logging.FrameworkDebug(ctx, "State updated due to semantic equality")

	resp.State.Raw = semanticEqualityResp.NewData.TerraformValue
}
