package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ReadResourceRequest is the framework server request for the
// ReadResource RPC.
type ReadResourceRequest struct {
	CurrentState *tfsdk.State
	ResourceType tfsdk.ResourceType
	Private      []byte
	ProviderMeta *tfsdk.Config
}

// ReadResourceResponse is the framework server response for the
// ReadResource RPC.
type ReadResourceResponse struct {
	Diagnostics diag.Diagnostics
	NewState    *tfsdk.State
	Private     []byte
}

// ReadResource implements the framework server ReadResource RPC.
func (s *Server) ReadResource(ctx context.Context, req *ReadResourceRequest, resp *ReadResourceResponse) {
	if req == nil {
		return
	}

	if req.CurrentState == nil {
		resp.Diagnostics.AddError(
			"Unexpected Read Request",
			"An unexpected error was encountered when reading the resource. The current state was missing.\n\n"+
				"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
		)

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

	readReq := tfsdk.ReadResourceRequest{
		State: tfsdk.State{
			Schema: req.CurrentState.Schema,
			Raw:    req.CurrentState.Raw.Copy(),
		},
	}
	readResp := tfsdk.ReadResourceResponse{
		State: tfsdk.State{
			Schema: req.CurrentState.Schema,
			Raw:    req.CurrentState.Raw.Copy(),
		},
	}

	if req.ProviderMeta != nil {
		readReq.ProviderMeta = *req.ProviderMeta
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Resource Read")
	resource.Read(ctx, readReq, &readResp)
	logging.FrameworkDebug(ctx, "Called provider defined Resource Read")

	resp.Diagnostics = readResp.Diagnostics
	resp.NewState = &readResp.State
}
