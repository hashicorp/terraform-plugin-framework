package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// DeleteResourceRequest is the framework server request for a delete request
// with the ApplyResourceChange RPC.
type DeleteResourceRequest struct {
	PlannedPrivate []byte
	PriorState     *tfsdk.State
	ProviderMeta   *tfsdk.Config
	ResourceSchema tfsdk.Schema
	ResourceType   tfsdk.ResourceType
}

// DeleteResourceResponse is the framework server response for a delete request
// with the ApplyResourceChange RPC.
type DeleteResourceResponse struct {
	Diagnostics diag.Diagnostics
	NewState    *tfsdk.State
	Private     []byte
}

// DeleteResource implements the framework server delete request logic for the
// ApplyResourceChange RPC.
func (s *Server) DeleteResource(ctx context.Context, req *DeleteResourceRequest, resp *DeleteResourceResponse) {
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

	deleteReq := tfsdk.DeleteResourceRequest{
		State: tfsdk.State{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
	}
	deleteResp := tfsdk.DeleteResourceResponse{
		State: tfsdk.State{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
	}

	if req.PriorState != nil {
		deleteReq.State = *req.PriorState
		deleteResp.State = *req.PriorState
	}

	if req.ProviderMeta != nil {
		deleteReq.ProviderMeta = *req.ProviderMeta
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Resource Delete")
	resource.Delete(ctx, deleteReq, &deleteResp)
	logging.FrameworkDebug(ctx, "Called provider defined Resource Delete")

	if !deleteResp.Diagnostics.HasError() {
		logging.FrameworkTrace(ctx, "No provider defined Delete errors detected, ensuring State is cleared")
		deleteResp.State.RemoveResource(ctx)
	}

	resp.Diagnostics = deleteResp.Diagnostics
	resp.NewState = &deleteResp.State
}
