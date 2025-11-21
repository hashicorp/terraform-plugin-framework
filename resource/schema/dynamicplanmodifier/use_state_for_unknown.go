// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamicplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateForUnknown returns a plan modifier that copies a known prior state
// value into the planned value. Use this when it is known that an unconfigured
// value will remain the same after a resource update.
//
// To prevent Terraform errors, the framework automatically sets unconfigured
// and Computed attributes to an unknown value "(known after apply)" on update.
// Using this plan modifier will instead display the prior state value in the
// plan, unless a prior plan modifier adjusts the value.
//
// Null is also a known value in Terraform and will be copied to the planned value
// by this plan modifier. For use-cases like a child attribute of a nested attribute or
// if null is desired to be marked as unknown in the case of an update, use [UseNonNullStateForUnknown].
func UseStateForUnknown() planmodifier.Dynamic {
	return useStateForUnknownModifier{}
}

// useStateForUnknownModifier implements the plan modifier.
type useStateForUnknownModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m useStateForUnknownModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModifyDynamic implements the plan modification logic.
func (m useStateForUnknownModifier) PlanModifyDynamic(ctx context.Context, req planmodifier.DynamicRequest, resp *planmodifier.DynamicResponse) {
	// Do nothing if there is no state (resource is being created).
	if req.State.Raw.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	// This also requires checking if the underlying value is known.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsUnderlyingValueUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	// This also requires checking if the underlying value is unknown.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsUnderlyingValueUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}
