// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int32planmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillBeBetween returns a plan modifier that will add a refinement to an unknown planned value
// which promises that:
//   - The final value will not be null.
//   - The final value will be greater than or equal to the provided minimum value.
//   - The final value will be less than or equal to the provided maximum value.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count".
func WillBeBetween(minVal, maxVal int32) planmodifier.Int32 {
	return willBeBetweenModifier{
		min: minVal,
		max: maxVal,
	}
}

type willBeBetweenModifier struct {
	min int32
	max int32
}

func (m willBeBetweenModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be between %d and %d once it becomes known", m.min, m.max)
}

func (m willBeBetweenModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be between %d and %d once it becomes known", m.min, m.max)
}

func (m willBeBetweenModifier) PlanModifyInt32(ctx context.Context, req planmodifier.Int32Request, resp *planmodifier.Int32Response) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.
		RefineWithLowerBound(m.min, true).
		RefineWithUpperBound(m.max, true)
}
