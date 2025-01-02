// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int64planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillBeAtMostModifierPlanModifyInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		maxVal   int64
		request  planmodifier.Int64Request
		expected *planmodifier.Int64Response
	}{
		"known-plan": {
			maxVal: 10,
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Value(5),
				PlanValue:   types.Int64Value(10),
				ConfigValue: types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Value(10),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown (with no refinement), otherwise they'll
			// get apply-time errors for changing the value even though
			// we knew it was legitimately possible for it to change and the
			// provider can't prevent this from happening
			maxVal: 10,
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Value(10),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Unknown(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown(),
			},
		},
		"unknown-plan-null-state": {
			maxVal: 10,
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Null(),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown().RefineWithUpperBound(10, true),
			},
		},
		"unknown-plan-non-null-state": {
			maxVal: 4,
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Value(10),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown().RefineWithUpperBound(4, true),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			maxVal: 6,
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Null(),
				PlanValue:   types.Int64Unknown().RefineWithLowerBound(2, false),
				ConfigValue: types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown().RefineWithLowerBound(2, false).RefineWithUpperBound(6, true),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Int64Response{
				PlanValue: testCase.request.PlanValue,
			}

			int64planmodifier.WillBeAtMost(testCase.maxVal).PlanModifyInt64(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
