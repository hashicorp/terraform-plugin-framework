// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillHaveSizeAtMost returns a plan modifier that will add a refinement to an unknown planned value
// which promises that:
//   - The final value will not be null.
//   - The final size of the list value will be at most the provided maximum value.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count".
func WillHaveSizeAtMost(maxVal int) planmodifier.List {
	return willHaveSizeAtMostModifier{
		max: maxVal,
	}
}

type willHaveSizeAtMostModifier struct {
	max int
}

func (m willHaveSizeAtMostModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will contain at most %d elements once it becomes known", m.max)
}

func (m willHaveSizeAtMostModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will contain at most %d elements once it becomes known", m.max)
}

func (m willHaveSizeAtMostModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineWithLengthUpperBound(int64(m.max))
}
