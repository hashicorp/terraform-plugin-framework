// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package setplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillHaveSizeBetween returns a plan modifier that will add a refinement to an unknown planned value
// which promises that:
//   - The final value will not be null.
//   - The final size of the set value will be at least the provided minimum value.
//   - The final size of the set value will be at most the provided maximum value.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count".
func WillHaveSizeBetween(minVal, maxVal int) planmodifier.Set {
	return willHaveSizeBetweenModifier{
		min: minVal,
		max: maxVal,
	}
}

type willHaveSizeBetweenModifier struct {
	min int
	max int
}

func (m willHaveSizeBetweenModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will contain at least %d elements and at most %d elements once it becomes known", m.min, m.max)
}

func (m willHaveSizeBetweenModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will contain at least %d elements and at most %d elements once it becomes known", m.min, m.max)
}

func (m willHaveSizeBetweenModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.
		RefineWithLengthLowerBound(int64(m.min)).
		RefineWithLengthUpperBound(int64(m.max))
}
