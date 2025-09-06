// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float64planmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateForUnknownIf returns a plan modifier that conditionally copies a known prior state
// value into the planned value. Use this when it is known that an unconfigured
// value will remain the same after a resource update, but only if the given
// condition is met.
//
// To prevent Terraform errors, the framework automatically sets unconfigured
// and Computed attributes to an unknown value "(known after apply)" on update.
// Using this plan modifier will instead display the prior state value in the
// plan, unless a prior plan modifier adjusts the value, but only if the
// condition function returns true.
func UseStateForUnknownIf(f UseStateForUnknownIfFunc, description, markdownDescription string) planmodifier.Float64 {
	return useStateForUnknownIfModifier{
		ifFunc:              f,
		description:         description,
		markdownDescription: markdownDescription,
	}
}

// useStateForUnknownIfModifier implements the conditional plan modifier.
type useStateForUnknownIfModifier struct {
	ifFunc              UseStateForUnknownIfFunc
	description         string
	markdownDescription string
}

// Description returns a human-readable description of the plan modifier.
func (m useStateForUnknownIfModifier) Description(_ context.Context) string {
	return m.description
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useStateForUnknownIfModifier) MarkdownDescription(_ context.Context) string {
	return m.markdownDescription
}

// PlanModifyFloat64 implements the plan modification logic.
func (m useStateForUnknownIfModifier) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	// Do nothing if there is no state (resource is being created).
	if req.State.Raw.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	ifFuncResp := &UseStateForUnknownIfFuncResponse{}

	m.ifFunc(ctx, req, ifFuncResp)

	resp.Diagnostics.Append(ifFuncResp.Diagnostics...)

	if ifFuncResp.UseState {
		resp.PlanValue = req.StateValue
	}
}
