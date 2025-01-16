// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package numberplanmodifier

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// WillBeAtMost returns a plan modifier that will add a refinement to an unknown planned value
// which promises that:
//   - The final value will not be null.
//   - The final value will be less than or equal to the provided maximum value.
//
// This unknown value refinement allows Terraform to validate more of the configuration during plan
// and evaluate conditional logic in meta-arguments such as "count".
func WillBeAtMost(maxVal *big.Float) planmodifier.Number {
	return willBeAtMostModifier{
		max: maxVal,
	}
}

type willBeAtMostModifier struct {
	max *big.Float
}

func (m willBeAtMostModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be at most %s once it becomes known", m.max.String())
}

func (m willBeAtMostModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value of this attribute will be at most %s once it becomes known", m.max.String())
}

func (m willBeAtMostModifier) PlanModifyNumber(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
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
