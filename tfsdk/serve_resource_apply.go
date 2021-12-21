package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

// IsCreate returns true if the request is creating a resource.
func serveResourceApplyIsCreate(ctx context.Context, state *Data) bool {
	// if our prior state isn't null, the state already exists, this can't
	// be a create request
	if !state.ReadOnlyData.Values.Null {
		return false
	}

	// otherwise, it's a create request
	return true
}

// IsUpdate returns true if the request is updating a resource.
func serveResourceApplyIsUpdate(ctx context.Context, state, plan *Data) bool {
	// if our prior state is null, the state doesn't exist, so this can't be
	// an update request
	if state.ReadOnlyData.Values.Null {
		return false
	}

	// if our planned state is null, this is a delete request, and it can't be
	// an update too
	if plan.ReadOnlyData.Values.Null {
		return false
	}

	// otherwise, this is an update
	return true
}

// IsDestroy returns true if the request is deleting a resource.
func serveResourceApplyIsDestroy(ctx context.Context, plan *Data) bool {
	// if our planned state isn't null, this can't be a delete request
	if !plan.ReadOnlyData.Values.Null {
		return false
	}

	// otherwise, this is a delete request
	return true
}

func serveResourceApplyCreate(ctx context.Context, resource Resource, config ReadOnlyData, plan *Data, usePM bool, pm ReadOnlyData, diags diag.Diagnostics) (*Data, diag.Diagnostics) {
	tfsdklog.Trace(ctx, "running create")
	req := CreateResourceRequest{
		Config: config,
		Plan:   plan,
	}
	if usePM {
		req.ProviderMeta = pm
	}
	resp := CreateResourceResponse{
		State:       state,
		Diagnostics: diags,
	}
	resource.Create(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		return nil, resp.Diagnostics
	}
	return resp.State, resp.Diagnostics
}

func serveResourceApplyUpdate(ctx context.Context, resource Resource, config ReadOnlyData, plan, state *Data, usePM bool, pm ReadOnlyData, diags diag.Diagnostics) (*Data, diag.Diagnostics) {
	tfsdklog.Trace(ctx, "running update")
	req := UpdateResourceRequest{
		Config: config,
		Plan:   plan,
		State:  state,
	}
	if usePM {
		req.ProviderMeta = pm
	}
	resp := UpdateResourceResponse{
		State:       state,
		Diagnostics: diags,
	}
	resource.Update(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		return nil, resp.Diagnostics
	}
	return resp.State, resp.Diagnostics
}

func serveResourceApplyDelete(ctx context.Context, resource Resource, state *Data, usePM bool, pm ReadOnlyData, diags diag.Diagnostics) (*Data, diag.Diagnostics) {
	tfsdklog.Trace(ctx, "running delete")
	req := DeleteResourceRequest{
		State: state,
	}
	if usePM {
		req.ProviderMeta = pm
	}
	resp := UpdateResourceResponse{
		State:       state,
		Diagnostics: diags,
	}
	resource.Delete(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		return nil, resp.Diagnostics
	}
	return resp.State, resp.Diagnostics
}
