// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package numberplanmodifier_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillBeAtLeastModifierPlanModifyNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		minVal   *big.Float
		request  planmodifier.NumberRequest
		expected *planmodifier.NumberResponse
	}{
		"known-plan": {
			minVal: big.NewFloat(5.5),
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberValue(big.NewFloat(5.5)),
				PlanValue:   types.NumberValue(big.NewFloat(10.1)),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberValue(big.NewFloat(10.1)),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown (with no refinement), otherwise they'll
			// get apply-time errors for changing the value even though
			// we knew it was legitimately possible for it to change and the
			// provider can't prevent this from happening
			minVal: big.NewFloat(5.5),
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberValue(big.NewFloat(10.1)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberUnknown(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
		"unknown-plan-null-state": {
			minVal: big.NewFloat(5.5),
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown().RefineWithLowerBound(big.NewFloat(5.5), true),
			},
		},
		"unknown-plan-non-null-state": {
			minVal: big.NewFloat(3.5),
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberValue(big.NewFloat(10.1)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown().RefineWithLowerBound(big.NewFloat(3.5), true),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			minVal: big.NewFloat(2.5),
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown().RefineWithUpperBound(big.NewFloat(6.1), false),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown().RefineWithUpperBound(big.NewFloat(6.1), false).RefineWithLowerBound(big.NewFloat(2.5), true),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.NumberResponse{
				PlanValue: testCase.request.PlanValue,
			}

			numberplanmodifier.WillBeAtLeast(testCase.minVal).PlanModifyNumber(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
