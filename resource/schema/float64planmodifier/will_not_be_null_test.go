// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float64planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillNotBeNullModifierPlanModifyFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Float64Request
		expected *planmodifier.Float64Response
	}{
		"known-plan": {
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Value(5),
				PlanValue:   types.Float64Value(10),
				ConfigValue: types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Value(10),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown (with no refinement), otherwise they'll
			// get apply-time errors for changing the value even though
			// we knew it was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Value(10),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Unknown(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown(),
			},
		},
		"unknown-plan-null-state": {
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Null(),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown().RefineAsNotNull(),
			},
		},
		"unknown-plan-non-null-state": {
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Value(10),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown().RefineAsNotNull(),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Null(),
				PlanValue:   types.Float64Unknown().RefineWithLowerBound(10, false),
				ConfigValue: types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown().RefineAsNotNull().RefineWithLowerBound(10, false),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Float64Response{
				PlanValue: testCase.request.PlanValue,
			}

			float64planmodifier.WillNotBeNull().PlanModifyFloat64(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
