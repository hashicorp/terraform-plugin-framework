// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ValidateStateStoreRequest is the framework server request for the
// ValidateStateStore RPC.
type ValidateStateStoreRequest struct {
	Config     *tfsdk.Config
	StateStore statestore.StateStore
}

// ValidateStateStoreResponse is the framework server response for the
// ValidateStateStore RPC.
type ValidateStateStoreResponse struct {
	Diagnostics diag.Diagnostics
}

// ValidateStateStore implements the framework server ValidateStateStore RPC.
func (s *Server) ValidateStateStore(ctx context.Context, req *ValidateStateStoreRequest, resp *ValidateStateStoreResponse) {
	if req == nil || req.Config == nil {
		return
	}

	if statestoreResourceWithConfigure, ok := req.StateStore.(statestore.StateStoreWithConfigure); ok {
		logging.FrameworkTrace(ctx, "StateStore implements StateStoreWithConfigure")

		configureReq := statestore.ConfigureStateStoreRequest{}
		configureResp := statestore.ConfigureStateStoreResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined StateStore Configure")
		statestoreResourceWithConfigure.Configure(ctx, configureReq, &configureResp)
		logging.FrameworkTrace(ctx, "Called provider defined StateStore Configure")

		resp.Diagnostics.Append(configureResp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	vdscReq := statestore.ValidateConfigRequest{
		Config: *req.Config,
	}

	if statestoreResourceWithConfigValidators, ok := req.StateStore.(statestore.StateStoreWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "StateStore implements StateStoreWithConfigValidators")

		for _, configValidator := range statestoreResourceWithConfigValidators.ConfigValidators(ctx) {
			// Instantiate a new response for each request to prevent validators
			// from modifying or removing diagnostics.
			vdscResp := &statestore.ValidateConfigResponse{}

			logging.FrameworkTrace(
				ctx,
				"Calling provider defined StateStoreValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)
			configValidator.ValidateStateStore(ctx, vdscReq, vdscResp)
			logging.FrameworkTrace(
				ctx,
				"Called provider defined StateStoreValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)

			resp.Diagnostics.Append(vdscResp.Diagnostics...)
		}
	}

	if statestoreResourceWithValidateConfig, ok := req.StateStore.(statestore.StateStoreWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "StateStore implements StateStoreWithValidateConfig")

		// Instantiate a new response for each request to prevent validators
		// from modifying or removing diagnostics.
		vdscResp := &statestore.ValidateConfigResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined StateStore ValidateConfig")
		statestoreResourceWithValidateConfig.ValidateConfig(ctx, vdscReq, vdscResp)
		logging.FrameworkTrace(ctx, "Called provider defined StateStore ValidateConfig")

		resp.Diagnostics.Append(vdscResp.Diagnostics...)
	}

	schemaCapabilities := validator.ValidateSchemaClientCapabilities{
		// The SchemaValidate function is shared between provider, resource,
		// data source and statestore resource schemas; however, WriteOnlyAttributesAllowed
		// capability is only valid for resource schemas, so this is explicitly set to false
		// for all other schema types.
		WriteOnlyAttributesAllowed: false,
	}

	validateSchemaReq := ValidateSchemaRequest{
		ClientCapabilities: schemaCapabilities,
		Config:             *req.Config,
	}
	// Instantiate a new response for each request to prevent validators
	// from modifying or removing diagnostics.
	validateSchemaResp := ValidateSchemaResponse{}

	SchemaValidate(ctx, req.Config.Schema, validateSchemaReq, &validateSchemaResp)

	resp.Diagnostics.Append(validateSchemaResp.Diagnostics...)
}
