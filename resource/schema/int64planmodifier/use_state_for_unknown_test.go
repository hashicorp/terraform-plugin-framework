// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int64planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUseStateForUnknownModifierPlanModifyInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Int64Request
		expected *planmodifier.Int64Response
	}{
		"null-state": {
			// when we first create the resource, use the unknown
			// value
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Null(),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Value(2),
				PlanValue:   types.Int64Value(1),
				ConfigValue: types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Value(1),
			},
		},
		"non-null-state-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Value(1),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Value(1),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.Int64Request{
				StateValue:  types.Int64Value(1),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Unknown(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown(),
			},
		},
		"under-list": {
			request: planmodifier.Int64Request{
				ConfigValue: types.Int64Null(),
				Path:        path.Root("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:   types.Int64Unknown(),
				StateValue:  types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
					),
				},
				PlanValue: types.Int64Unknown(),
			},
		},
		"under-set": {
			request: planmodifier.Int64Request{
				ConfigValue: types.Int64Null(),
				Path: path.Root("test").AtSetValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_test": types.Int64Type,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_test": types.Int64Type,
								},
								map[string]attr.Value{
									"nested_test": types.Int64Unknown(),
								},
							),
						},
					),
				).AtName("nested_test"),
				PlanValue:  types.Int64Unknown(),
				StateValue: types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.SetValueMust(
								types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_test": types.Int64Type,
									},
								},
								[]attr.Value{
									types.ObjectValueMust(
										map[string]attr.Type{
											"nested_test": types.Int64Type,
										},
										map[string]attr.Value{
											"nested_test": types.Int64Unknown(),
										},
									),
								},
							),
						).AtName("nested_test"),
					),
				},
				PlanValue: types.Int64Unknown(),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Int64Response{
				PlanValue: testCase.request.PlanValue,
			}

			int64planmodifier.UseStateForUnknown().PlanModifyInt64(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
