// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillHaveSizeAtLeastModifierPlanModifyMap(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		minVal   int
		request  planmodifier.MapRequest
		expected *planmodifier.MapResponse
	}{
		"known-plan": {
			minVal: 5,
			request: planmodifier.MapRequest{
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("hello"), "key2": types.StringValue("world")}),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("hello"), "key2": types.StringValue("world"), "key3": types.StringValue("!")}),
				ConfigValue: types.MapNull(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("hello"), "key2": types.StringValue("world"), "key3": types.StringValue("!")}),
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
			request: planmodifier.MapRequest{
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("hello"), "key2": types.StringValue("world"), "key3": types.StringValue("!")}),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapUnknown(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapUnknown(types.StringType),
			},
		},
		"unknown-plan-null-state": {
			minVal: 5,
			request: planmodifier.MapRequest{
				StateValue:  types.MapNull(types.StringType),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapNull(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapUnknown(types.StringType).RefineWithLengthLowerBound(5),
			},
		},
		"unknown-plan-non-null-state": {
			minVal: 3,
			request: planmodifier.MapRequest{
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("hello"), "key2": types.StringValue("world"), "key3": types.StringValue("!")}),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapNull(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapUnknown(types.StringType).RefineWithLengthLowerBound(3),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			minVal: 2,
			request: planmodifier.MapRequest{
				StateValue:  types.MapNull(types.StringType),
				PlanValue:   types.MapUnknown(types.StringType).RefineWithLengthUpperBound(6),
				ConfigValue: types.MapNull(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapUnknown(types.StringType).RefineWithLengthUpperBound(6).RefineWithLengthLowerBound(2),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.MapResponse{
				PlanValue: testCase.request.PlanValue,
			}

			mapplanmodifier.WillHaveSizeAtLeast(testCase.minVal).PlanModifyMap(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
