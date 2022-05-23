package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// CreateResourceRequest is the framework server request for a create request
// with the ApplyResourceChange RPC.
type CreateResourceRequest struct {
	Config         *tfsdk.Config
	PlannedPrivate []byte
	PlannedState   *tfsdk.Plan
	ProviderMeta   *tfsdk.Config
	ResourceSchema tfsdk.Schema
	ResourceType   tfsdk.ResourceType
}

// CreateResourceResponse is the framework server response for a create request
// with the ApplyResourceChange RPC.
type CreateResourceResponse struct {
	Diagnostics diag.Diagnostics
	NewState    *tfsdk.State
	Private     []byte
}

// CreateResource implements the framework server create request logic for the
// ApplyResourceChange RPC.
func (s *Server) CreateResource(ctx context.Context, req *CreateResourceRequest, resp *CreateResourceResponse) {
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

	createReq := tfsdk.CreateResourceRequest{
		Config: tfsdk.Config{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
		Plan: tfsdk.Plan{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
	}
	createResp := tfsdk.CreateResourceResponse{
		State: tfsdk.State{
			Schema: req.ResourceSchema,
			Raw:    tftypes.NewValue(req.ResourceSchema.TerraformType(ctx), nil),
		},
	}

	if req.Config != nil {
		createReq.Config = *req.Config
	}

	if req.PlannedState != nil {
		createReq.Plan = *req.PlannedState
	}

	if req.ProviderMeta != nil {
		createReq.ProviderMeta = *req.ProviderMeta
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Resource Create")
	resource.Create(ctx, createReq, &createResp)
	logging.FrameworkDebug(ctx, "Called provider defined Resource Create")

	resp.Diagnostics = createResp.Diagnostics
	resp.NewState = &createResp.State
}
