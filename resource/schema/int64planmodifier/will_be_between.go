// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int64planmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// TODO: docs
func WillBeBetween(minVal, maxVal int64) planmodifier.Int64 {
	return willBeBetweenModifier{
		min: minVal,
		max: maxVal,
	}
}

type willBeBetweenModifier struct {
	min int64
	max int64
}

func (m willBeBetweenModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value will be between %d and %d once it becomes known", m.min, m.max)
}

func (m willBeBetweenModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value will be between %d and %d once it becomes known", m.min, m.max)
}

func (m willBeBetweenModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
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
