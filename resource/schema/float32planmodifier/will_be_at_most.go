// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float32planmodifier

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
func WillBeAtMost(maxVal float32) planmodifier.Float32 {
	return willBeAtMostModifier{
		max: maxVal,
	}
}

type willBeAtMostModifier struct {
	max float32
}

func (m willBeAtMostModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be at most %f once it becomes known", m.max)
}

func (m willBeAtMostModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be at most %f once it becomes known", m.max)
}

func (m willBeAtMostModifier) PlanModifyFloat32(ctx context.Context, req planmodifier.Float32Request, resp *planmodifier.Float32Response) {
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
