// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int32planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillBeBetweenModifierPlanModifyInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		minVal   int32
		maxVal   int32
		request  planmodifier.Int32Request
		expected *planmodifier.Int32Response
	}{
		"known-plan": {
			minVal: 5,
			maxVal: 10,
			request: planmodifier.Int32Request{
				StateValue:  types.Int32Value(5),
				PlanValue:   types.Int32Value(10),
				ConfigValue: types.Int32Null(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Value(10),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown (with no refinement), otherwise they'll
			// get apply-time errors for changing the value even though
			// we knew it was legitimately possible for it to change and the
			// provider can't prevent this from happening
			minVal: 5,
			maxVal: 10,
			request: planmodifier.Int32Request{
				StateValue:  types.Int32Value(10),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Unknown(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown(),
			},
		},
		"unknown-plan-null-state": {
			minVal: 5,
			maxVal: 10,
			request: planmodifier.Int32Request{
				StateValue:  types.Int32Null(),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown().RefineWithLowerBound(5, true).RefineWithUpperBound(10, true),
			},
		},
		"unknown-plan-non-null-state": {
			minVal: 3,
			maxVal: 4,
			request: planmodifier.Int32Request{
				StateValue:  types.Int32Value(10),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown().RefineWithLowerBound(3, true).RefineWithUpperBound(4, true),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			minVal: 2,
			maxVal: 6,
			request: planmodifier.Int32Request{
				StateValue:  types.Int32Null(),
				PlanValue:   types.Int32Unknown().RefineAsNotNull(),
				ConfigValue: types.Int32Null(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown().RefineAsNotNull().RefineWithLowerBound(2, true).RefineWithUpperBound(6, true),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Int32Response{
				PlanValue: testCase.request.PlanValue,
			}

			int32planmodifier.WillBeBetween(testCase.minVal, testCase.maxVal).PlanModifyInt32(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
