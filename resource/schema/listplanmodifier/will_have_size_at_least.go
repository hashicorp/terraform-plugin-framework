// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillHaveSizeAtLeast returns a plan modifier that will add a refinement to an unknown planned value
// which promises that:
//   - The final value will not be null.
//   - The final size of the list value will be at least the provided minimum value.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count".
func WillHaveSizeAtLeast(minVal int) planmodifier.List {
	return willHaveSizeAtLeastModifier{
		min: minVal,
	}
}

type willHaveSizeAtLeastModifier struct {
	min int
}

func (m willHaveSizeAtLeastModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will contain at least %d elements once it becomes known", m.min)
}

func (m willHaveSizeAtLeastModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will contain at least %d elements once it becomes known", m.min)
}

func (m willHaveSizeAtLeastModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineWithLengthLowerBound(int64(m.min))
}
