// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stringplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillNotBeNullModifierPlanModifyString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.StringRequest
		expected *planmodifier.StringResponse
	}{
		"known-plan": {
			request: planmodifier.StringRequest{
				StateValue:  types.StringValue("other"),
				PlanValue:   types.StringValue("test"),
				ConfigValue: types.StringNull(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringValue("test"),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown (with no refinement), otherwise they'll
			// get apply-time errors for changing the value even though
			// we knew it was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.StringRequest{
				StateValue:  types.StringValue("test"),
				PlanValue:   types.StringUnknown(),
				ConfigValue: types.StringUnknown(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringUnknown(),
			},
		},
		"unknown-plan-null-state": {
			request: planmodifier.StringRequest{
				StateValue:  types.StringNull(),
				PlanValue:   types.StringUnknown(),
				ConfigValue: types.StringNull(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringUnknown().RefineAsNotNull(),
			},
		},
		"unknown-plan-non-null-state": {
			request: planmodifier.StringRequest{
				StateValue:  types.StringValue("test"),
				PlanValue:   types.StringUnknown(),
				ConfigValue: types.StringNull(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringUnknown().RefineAsNotNull(),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			request: planmodifier.StringRequest{
				StateValue:  types.StringNull(),
				PlanValue:   types.StringUnknown().RefineWithPrefix("preserve me"),
				ConfigValue: types.StringNull(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue: types.StringUnknown().RefineAsNotNull().RefineWithPrefix("preserve me"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.StringResponse{
				PlanValue: testCase.request.PlanValue,
			}

			stringplanmodifier.WillNotBeNull().PlanModifyString(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
