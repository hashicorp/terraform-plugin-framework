package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// UpdateResourceRequest is the framework server request for an update request
// with the ApplyResourceChange RPC.
type UpdateResourceRequest struct {
	Config         *tfsdk.Config
	PlannedPrivate []byte
	PlannedState   *tfsdk.Plan
	PriorState     *tfsdk.State
	ProviderMeta   *tfsdk.Config
	ResourceSchema tfsdk.Schema
	ResourceType   tfsdk.ResourceType
}

// UpdateResourceResponse is the framework server response for an update request
// with the ApplyResourceChange RPC.
type UpdateResourceResponse struct {
	Diagnostics diag.Diagnostics
	NewState    *tfsdk.State
	Private     []byte
}

// UpdateResource implements the framework server update request logic for the
// ApplyResourceChange RPC.
func (s *Server) UpdateResource(ctx context.Context, req *UpdateResourceRequest, resp *UpdateResourceResponse) {
	if req == nil {
		return
	}

	// Always instantiate new Resource instances.
	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := req.ResourceType.NewResource(ctx, s.Provider)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := tfsdk.UpdateResourceRequest{
		Config: tfsdk.Config{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
		Plan: tfsdk.Plan{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
		State: tfsdk.State{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
	}
	updateResp := tfsdk.UpdateResourceResponse{
		State: tfsdk.State{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
	}

	if req.Config != nil {
		updateReq.Config = *req.Config
	}

	if req.PlannedState != nil {
		updateReq.Plan = *req.PlannedState
	}

	if req.PriorState != nil {
		updateReq.State = *req.PriorState
		// Require explicit provider updates for tracking successful updates.
		updateResp.State = *req.PriorState
	}

	if req.ProviderMeta != nil {
		updateReq.ProviderMeta = *req.ProviderMeta
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Resource Update")
	resource.Update(ctx, updateReq, &updateResp)
	logging.FrameworkDebug(ctx, "Called provider defined Resource Update")

	resp.Diagnostics = updateResp.Diagnostics
	resp.NewState = &updateResp.State
}
