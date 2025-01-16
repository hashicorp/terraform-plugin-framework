// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float32planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillBeAtMostModifierPlanModifyFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		maxVal   float32
		request  planmodifier.Float32Request
		expected *planmodifier.Float32Response
	}{
		"known-plan": {
			maxVal: 10.1,
			request: planmodifier.Float32Request{
				StateValue:  types.Float32Value(5),
				PlanValue:   types.Float32Value(10),
				ConfigValue: types.Float32Null(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Value(10),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown (with no refinement), otherwise they'll
			// get apply-time errors for changing the value even though
			// we knew it was legitimately possible for it to change and the
			// provider can't prevent this from happening
			maxVal: 10.1,
			request: planmodifier.Float32Request{
				StateValue:  types.Float32Value(10),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Unknown(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown(),
			},
		},
		"unknown-plan-null-state": {
			maxVal: 10.1,
			request: planmodifier.Float32Request{
				StateValue:  types.Float32Null(),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown().RefineWithUpperBound(10.1, true),
			},
		},
		"unknown-plan-non-null-state": {
			maxVal: 4.1,
			request: planmodifier.Float32Request{
				StateValue:  types.Float32Value(10),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown().RefineWithUpperBound(4.1, true),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			maxVal: 6.1,
			request: planmodifier.Float32Request{
				StateValue:  types.Float32Null(),
				PlanValue:   types.Float32Unknown().RefineWithLowerBound(2, false),
				ConfigValue: types.Float32Null(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown().RefineWithLowerBound(2, false).RefineWithUpperBound(6.1, true),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Float32Response{
				PlanValue: testCase.request.PlanValue,
			}

			float32planmodifier.WillBeAtMost(testCase.maxVal).PlanModifyFloat32(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
