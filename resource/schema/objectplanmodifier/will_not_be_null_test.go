// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package objectplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWillNotBeNullModifierPlanModifyObject(t *testing.T) {
	t.Parallel()

	objType := map[string]attr.Type{
		"attr_one": types.StringType,
	}

	testCases := map[string]struct {
		request  planmodifier.ObjectRequest
		expected *planmodifier.ObjectResponse
	}{
		"known-plan": {
			request: planmodifier.ObjectRequest{
				StateValue: types.ObjectValueMust(objType, map[string]attr.Value{
					"attr_one": types.StringValue("hello!"),
				}),
				PlanValue: types.ObjectValueMust(objType, map[string]attr.Value{
					"attr_one": types.StringValue("world!"),
				}),
				ConfigValue: types.ObjectNull(objType),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectValueMust(objType, map[string]attr.Value{
					"attr_one": types.StringValue("world!"),
				}),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown (with no refinement), otherwise they'll
			// get apply-time errors for changing the value even though
			// we knew it was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.ObjectRequest{
				StateValue: types.ObjectValueMust(objType, map[string]attr.Value{
					"attr_one": types.StringValue("world!"),
				}),
				PlanValue:   types.ObjectUnknown(objType),
				ConfigValue: types.ObjectUnknown(objType),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(objType),
			},
		},
		"unknown-plan-null-state": {
			request: planmodifier.ObjectRequest{
				StateValue:  types.ObjectNull(objType),
				PlanValue:   types.ObjectUnknown(objType),
				ConfigValue: types.ObjectNull(objType),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(objType).RefineAsNotNull(),
			},
		},
		"unknown-plan-non-null-state": {
			request: planmodifier.ObjectRequest{
				StateValue: types.ObjectValueMust(objType, map[string]attr.Value{
					"attr_one": types.StringValue("world!"),
				}),
				PlanValue:   types.ObjectUnknown(objType),
				ConfigValue: types.ObjectNull(objType),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(objType).RefineAsNotNull(),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.ObjectResponse{
				PlanValue: testCase.request.PlanValue,
			}

			objectplanmodifier.WillNotBeNull().PlanModifyObject(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
