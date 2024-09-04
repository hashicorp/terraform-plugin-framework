// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// CloseEphemeralResourceRequest is the framework server request for the
// CloseEphemeralResource RPC.
type CloseEphemeralResourceRequest struct {
	State                   *tfsdk.EphemeralState
	Private                 *privatestate.Data
	EphemeralResourceSchema fwschema.Schema
	EphemeralResource       ephemeral.EphemeralResource
}

// CloseEphemeralResourceResponse is the framework server response for the
// CloseEphemeralResource RPC.
type CloseEphemeralResourceResponse struct {
	Diagnostics diag.Diagnostics
}

// CloseEphemeralResource implements the framework server CloseEphemeralResource RPC.
func (s *Server) CloseEphemeralResource(ctx context.Context, req *CloseEphemeralResourceRequest, resp *CloseEphemeralResourceResponse) {
	if req == nil {
		return
	}

	if ephemeralResourceWithConfigure, ok := req.EphemeralResource.(ephemeral.EphemeralResourceWithConfigure); ok {
		logging.FrameworkTrace(ctx, "EphemeralResource implements EphemeralResourceWithConfigure")

		configureReq := ephemeral.ConfigureRequest{
			ProviderData: s.EphemeralResourceConfigureData,
		}
		configureResp := ephemeral.ConfigureResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined EphemeralResource Configure")
		ephemeralResourceWithConfigure.Configure(ctx, configureReq, &configureResp)
		logging.FrameworkTrace(ctx, "Called provider defined EphemeralResource Configure")

		resp.Diagnostics.Append(configureResp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	resourceWithClose, ok := req.EphemeralResource.(ephemeral.EphemeralResourceWithClose)
	if !ok {
		// The framework automatically sets the indicator to Terraform core that enables calling close using
		// this interface, so if we get this diagnostic then there is a bug in either Terraform core or framework.
		resp.Diagnostics.AddError(
			"Ephemeral Resource Close Not Implemented",
			"An unexpected error was encountered when closing the ephemeral resource. Terraform sent a close request for an "+
				"ephemeral resource that has not implemented close logic.\n\n"+
				"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
		)
		return
	}

	if req.State == nil {
		resp.Diagnostics.AddError(
			"Unexpected Close Request",
			"An unexpected error was encountered when closing the ephemeral resource. The state was missing.\n\n"+
				"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
		)

		return
	}

	privateProviderData := privatestate.EmptyProviderData(ctx)
	if req.Private != nil && req.Private.Provider != nil {
		privateProviderData = req.Private.Provider
	}

	closeReq := ephemeral.CloseRequest{
		State: tfsdk.EphemeralState{
			Schema: req.EphemeralResourceSchema,
			Raw:    req.State.Raw.Copy(),
		},
		Private: privateProviderData,
	}
	closeResp := ephemeral.CloseResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined EphemeralResource Close")
	resourceWithClose.Close(ctx, closeReq, &closeResp)
	logging.FrameworkTrace(ctx, "Called provider defined EphemeralResource Close")

	resp.Diagnostics = closeResp.Diagnostics
}
