// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package boolplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/parentpath"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
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
// To prevent data issues and Terraform errors, this plan modifier cannot be
// implemented on attribute values beneath lists or sets. An implementation
// error diagnostic is raised if the plan modifier logic detects a list or set
// in the request path.
func UseStateForUnknown() planmodifier.Bool {
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

// PlanModifyBool implements the plan modification logic.
func (m useStateForUnknownModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// Verify this plan modifier is not being used beneath a list or set.
	// Lists and sets do not have a generic methodology to identify/track
	// an element if rearranged, especially within an object with multiple
	// computed attribute values. Only the provider can determine which
	// underlying values in an element are significant to realign a prior
	// state value during updates.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/709
	if parentpath.HasListOrSet(req.Path) {
		resp.Diagnostics.Append(planmodifierdiag.UseStateForUnknownUnderListOrSet(req.Path))

		return
	}

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
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

	resp.PlanValue = req.StateValue
}
