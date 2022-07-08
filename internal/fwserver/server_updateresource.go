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

	nullSchemaData := tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil)

	updateReq := tfsdk.UpdateResourceRequest{
		Config: tfsdk.Config{
			Schema: req.ResourceSchema,
			Raw:    nullSchemaData,
		},
		Plan: tfsdk.Plan{
			Schema: req.ResourceSchema,
			Raw:    nullSchemaData,
		},
		State: tfsdk.State{
			Schema: req.ResourceSchema,
			Raw:    nullSchemaData,
		},
	}
	updateResp := tfsdk.UpdateResourceResponse{
		State: tfsdk.State{
			Schema: req.ResourceSchema,
			Raw:    nullSchemaData,
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

	if !resp.Diagnostics.HasError() && updateResp.State.Raw.Equal(nullSchemaData) {
		resp.Diagnostics.AddError(
			"Missing Resource State After Update",
			"The Terraform Provider unexpectedly returned no resource state after having no errors in the resource update. "+
				"This is always an issue in the Terraform Provider and should be reported to the provider developers.",
		)
	}
}
