// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float64planmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillBeAtMost returns a plan modifier that will add a refinement to an unknown planned value
// which promises that:
//   - The final value will not be null.
//   - The final value will be less than or equal to the provided maximum value.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count".
func WillBeAtMost(maxVal float64) planmodifier.Float64 {
	return willBeAtMostModifier{
		max: maxVal,
	}
}

type willBeAtMostModifier struct {
	max float64
}

func (m willBeAtMostModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be at most %f once it becomes known", m.max)
}

func (m willBeAtMostModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be at most %f once it becomes known", m.max)
}

func (m willBeAtMostModifier) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineWithUpperBound(m.max, true)
}
