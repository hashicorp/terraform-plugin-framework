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

func TestWillNotBeNullModifierPlanModifyNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.NumberRequest
		expected *planmodifier.NumberResponse
	}{
		"known-plan": {
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
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown().RefineAsNotNull(),
			},
		},
		"unknown-plan-non-null-state": {
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberValue(big.NewFloat(10.1)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown().RefineAsNotNull(),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown().RefineWithLowerBound(big.NewFloat(10.1), false),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown().RefineAsNotNull().RefineWithLowerBound(big.NewFloat(10.1), false),
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

			numberplanmodifier.WillNotBeNull().PlanModifyNumber(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
