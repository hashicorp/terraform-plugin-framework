// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillHaveSizeAtMostModifierPlanModifyList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		maxVal   int
		request  planmodifier.ListRequest
		expected *planmodifier.ListResponse
	}{
		"known-plan": {
			maxVal: 10,
			request: planmodifier.ListRequest{
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("hello"), types.StringValue("world")}),
				PlanValue:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("hello"), types.StringValue("world"), types.StringValue("!")}),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("hello"), types.StringValue("world"), types.StringValue("!")}),
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
			request: planmodifier.ListRequest{
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("hello"), types.StringValue("world"), types.StringValue("!")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListUnknown(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"unknown-plan-null-state": {
			maxVal: 10,
			request: planmodifier.ListRequest{
				StateValue:  types.ListNull(types.StringType),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType).RefineWithLengthUpperBound(10),
			},
		},
		"unknown-plan-non-null-state": {
			maxVal: 4,
			request: planmodifier.ListRequest{
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("hello"), types.StringValue("world"), types.StringValue("!")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType).RefineWithLengthUpperBound(4),
			},
		},
		"unknown-plan-preserve-existing-refinement": {
			maxVal: 6,
			request: planmodifier.ListRequest{
				StateValue:  types.ListNull(types.StringType),
				PlanValue:   types.ListUnknown(types.StringType).RefineWithLengthLowerBound(2),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType).RefineWithLengthLowerBound(2).RefineWithLengthUpperBound(6),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.ListResponse{
				PlanValue: testCase.request.PlanValue,
			}

			listplanmodifier.WillHaveSizeAtMost(testCase.maxVal).PlanModifyList(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
