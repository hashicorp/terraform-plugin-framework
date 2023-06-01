// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUseStateForUnknownModifierPlanModifyList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.ListRequest
		expected *planmodifier.ListResponse
	}{
		"null-state": {
			// when we first create the resource, use the unknown
			// value
			request: planmodifier.ListRequest{
				StateValue:  types.ListNull(types.StringType),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.ListRequest{
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				PlanValue:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
		},
		"non-null-state-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			request: planmodifier.ListRequest{
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.ListRequest{
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListUnknown(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"under-list": {
			request: planmodifier.ListRequest{
				ConfigValue: types.ListNull(types.StringType),
				Path:        path.Root("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:   types.ListUnknown(types.StringType),
				StateValue:  types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
					),
				},
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"under-set": {
			request: planmodifier.ListRequest{
				ConfigValue: types.ListNull(types.StringType),
				Path: path.Root("test").AtSetValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_test": types.ListType{ElemType: types.StringType},
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_test": types.ListType{ElemType: types.StringType},
								},
								map[string]attr.Value{
									"nested_test": types.ListUnknown(types.StringType),
								},
							),
						},
					),
				).AtName("nested_test"),
				PlanValue:  types.ListUnknown(types.StringType),
				StateValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.SetValueMust(
								types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_test": types.ListType{ElemType: types.StringType},
									},
								},
								[]attr.Value{
									types.ObjectValueMust(
										map[string]attr.Type{
											"nested_test": types.ListType{ElemType: types.StringType},
										},
										map[string]attr.Value{
											"nested_test": types.ListUnknown(types.StringType),
										},
									),
								},
							),
						).AtName("nested_test"),
					),
				},
				PlanValue: types.ListUnknown(types.StringType),
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

			listplanmodifier.UseStateForUnknown().PlanModifyList(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
