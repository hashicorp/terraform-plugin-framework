// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float32planmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillBeAtLeast returns a plan modifier that will add a refinement to an unknown planned value
// which promises that:
//   - The final value will not be null.
//   - The final value will be greater than or equal to the provided minimum value.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count".
func WillBeAtLeast(minVal float32) planmodifier.Float32 {
	return willBeAtLeastModifier{
		min: minVal,
	}
}

type willBeAtLeastModifier struct {
	min float32
}

func (m willBeAtLeastModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be at least %f once it becomes known", m.min)
}

func (m willBeAtLeastModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be at least %f once it becomes known", m.min)
}

func (m willBeAtLeastModifier) PlanModifyFloat32(ctx context.Context, req planmodifier.Float32Request, resp *planmodifier.Float32Response) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineWithLowerBound(m.min, true)
}
