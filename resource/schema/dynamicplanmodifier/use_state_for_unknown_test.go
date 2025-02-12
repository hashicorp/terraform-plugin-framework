// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamicplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUseStateForUnknownModifierPlanModifyDynamic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.DynamicRequest
		expected *planmodifier.DynamicResponse
	}{
		"null-state": {
			// when we first create the resource, use the unknown value
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicNull(),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicUnknown(),
			},
		},
		"null-underlying-state-value": {
			// if the state value has a known underlying type, but a null underlying value,
			// use the unknown value
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicValue(types.StringNull()),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicUnknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it. We still want to preserve that value, in this case
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicValue(types.StringValue("other")),
				PlanValue:   types.DynamicValue(types.StringValue("test")),
				ConfigValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringValue("test")),
			},
		},
		"known-plan-null": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it. We still want to preserve that value, in this case
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicValue(types.StringValue("other")),
				PlanValue:   types.DynamicNull(),
				ConfigValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicNull(),
			},
		},
		"known-underlying-plan-value-null": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it. We still want to preserve that value, in this case
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicValue(types.StringValue("other")),
				PlanValue:   types.DynamicValue(types.StringNull()),
				ConfigValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringNull()),
			},
		},
		"non-null-state-unknown-plan": {
			// this is the situation we want to preserve the state in
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicValue(types.StringValue("test")),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringValue("test")),
			},
		},
		"non-null-state-unknown-underlying-plan-value": {
			// if the plan value has a known underlying type, but an unknown underlying value
			// we want to preserve the state
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicValue(types.StringValue("test")),
				PlanValue:   types.DynamicValue(types.StringUnknown()),
				ConfigValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringValue("test")),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicValue(types.StringValue("test")),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicUnknown(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicUnknown(),
			},
		},
		"unknown-underlying-config-value": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.DynamicRequest{
				StateValue:  types.DynamicValue(types.StringValue("test")),
				PlanValue:   types.DynamicValue(types.StringUnknown()),
				ConfigValue: types.DynamicValue(types.StringUnknown()),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringUnknown()),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.DynamicResponse{
				PlanValue: testCase.request.PlanValue,
			}

			dynamicplanmodifier.UseStateForUnknown().PlanModifyDynamic(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
