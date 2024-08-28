// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// RenewEphemeralResourceRequest is the framework server request for the
// RenewEphemeralResource RPC.
type RenewEphemeralResourceRequest struct {
	Config                  *tfsdk.Config
	State                   *tfsdk.EphemeralState
	Private                 *privatestate.Data
	EphemeralResourceSchema fwschema.Schema
	EphemeralResource       ephemeral.EphemeralResource
}

// RenewEphemeralResourceResponse is the framework server response for the
// RenewEphemeralResource RPC.
type RenewEphemeralResourceResponse struct {
	Private     *privatestate.Data
	Diagnostics diag.Diagnostics
	RenewAt     time.Time
}

// RenewEphemeralResource implements the framework server RenewEphemeralResource RPC.
func (s *Server) RenewEphemeralResource(ctx context.Context, req *RenewEphemeralResourceRequest, resp *RenewEphemeralResourceResponse) {
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

	resourceWithRenew, ok := req.EphemeralResource.(ephemeral.EphemeralResourceWithRenew)
	if !ok {
		// TODO: this diagnostic should be more worded towards an invalid implementation or core bug.
		// Either something is bugged in core and called renew incorrectly, or the provider populated
		// the RenewAt response field without defining the renew function (invalid implementation).
		resp.Diagnostics.AddError(
			"Ephemeral Resource Renew Not Implemented",
			"This resource does not support renew. Please contact the provider developer for additional information.",
		)
		return
	}

	// Ensure that resp.Private is never nil.
	resp.Private = privatestate.EmptyData(ctx)
	if req.Private != nil {
		// Overwrite resp.Private with req.Private providing it is not nil.
		resp.Private = req.Private

		// Ensure that resp.Private.Provider is never nil.
		if resp.Private.Provider == nil {
			resp.Private.Provider = privatestate.EmptyProviderData(ctx)
		}
	}

	renewReq := ephemeral.RenewRequest{
		Config: tfsdk.Config{
			Schema: req.EphemeralResourceSchema,
		},
		State: tfsdk.EphemeralState{
			Schema: req.EphemeralResourceSchema,
			Raw:    req.State.Raw.Copy(),
		},
		Private: resp.Private.Provider,
	}
	renewResp := ephemeral.RenewResponse{
		Private: renewReq.Private,
	}

	if req.Config != nil {
		renewReq.Config = *req.Config
	}

	logging.FrameworkTrace(ctx, "Calling provider defined EphemeralResource Renew")
	resourceWithRenew.Renew(ctx, renewReq, &renewResp)
	logging.FrameworkTrace(ctx, "Called provider defined EphemeralResource Renew")

	resp.Diagnostics = renewResp.Diagnostics
	resp.RenewAt = renewResp.RenewAt

	if renewResp.Private != nil {
		resp.Private.Provider = renewResp.Private
	}
}
