// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ConfigureStateStoreRequest is the framework server request for the
// ConfigureStateStore RPC.
type ConfigureStateStoreRequest struct {
	Config             *tfsdk.Config
	StateStore         statestore.StateStore
	StateStoreSchema   fwschema.Schema
	ClientCapabilities statestore.ConfigureStateStoreClientCapabilities
}

// ConfigureStateStoreResponse is the framework server response for the
// ConfigureStateStore RPC.
type ConfigureStateStoreResponse struct {
	Diagnostics        diag.Diagnostics
	ServerCapabilities *StateStoreServerCapabilities
}

type stateStoreConfigureData struct {
	ServerCapabilities      StateStoreServerCapabilities
	StateStoreConfigureData any
}

// ConfigureStateStore implements the framework server ConfigureStateStore RPC.
func (s *Server) ConfigureStateStore(ctx context.Context, req *ConfigureStateStoreRequest, resp *ConfigureStateStoreResponse) {
	if req == nil {
		return
	}

	nullSchemaData := tftypes.NewValue(req.StateStoreSchema.Type().TerraformType(ctx), nil)
	configureReq := statestore.ConfigureStateStoreRequest{
		Config: tfsdk.Config{
			Schema: req.StateStoreSchema,
			Raw:    nullSchemaData,
		},
		ProviderData: s.StateStoreProviderData,
	}
	if req.Config != nil {
		configureReq.Config = *req.Config
	}

	configureResp := statestore.ConfigureStateStoreResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined StateStore ConfigureStateStore")
	req.StateStore.ConfigureStateStore(ctx, configureReq, &configureResp)
	logging.FrameworkTrace(ctx, "Called provider defined StateStore ConfigureStateStore")

	resp.Diagnostics = configureResp.Diagnostics
	resp.ServerCapabilities = &StateStoreServerCapabilities{
		// MAINTAINER NOTE: Currently, we just round-trip the proposed chunk size from Terraform core (8 MB). In the future,
		// we could expose this to provider developers in [statestore.ConfigureStateStoreResponse] if controlling
		// the chunk size is desired.
		ChunkSize: req.ClientCapabilities.ChunkSize,
	}

	// Set state store configure data + server capabilities for reference in future state store RPCs
	s.stateStoreConfigureData = stateStoreConfigureData{
		ServerCapabilities:      *resp.ServerCapabilities,
		StateStoreConfigureData: configureResp.StateStoreData,
	}
}
