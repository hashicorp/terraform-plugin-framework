// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float64planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUseStateForUnknownModifierPlanModifyFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Float64Request
		expected *planmodifier.Float64Response
	}{
		"null-state": {
			// when we first create the resource, use the unknown
			// value
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Null(),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Value(2.4),
				PlanValue:   types.Float64Value(1.2),
				ConfigValue: types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Value(1.2),
			},
		},
		"non-null-state-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Value(1.2),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Value(1.2),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.Float64Request{
				StateValue:  types.Float64Value(1.2),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Unknown(),
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown(),
			},
		},
		"under-list": {
			request: planmodifier.Float64Request{
				ConfigValue: types.Float64Null(),
				Path:        path.Root("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:   types.Float64Unknown(),
				StateValue:  types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
					),
				},
				PlanValue: types.Float64Unknown(),
			},
		},
		"under-set": {
			request: planmodifier.Float64Request{
				ConfigValue: types.Float64Null(),
				Path: path.Root("test").AtSetValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_test": types.Float64Type,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_test": types.Float64Type,
								},
								map[string]attr.Value{
									"nested_test": types.Float64Unknown(),
								},
							),
						},
					),
				).AtName("nested_test"),
				PlanValue:  types.Float64Unknown(),
				StateValue: types.Float64Null(),
			},
			expected: &planmodifier.Float64Response{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.SetValueMust(
								types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_test": types.Float64Type,
									},
								},
								[]attr.Value{
									types.ObjectValueMust(
										map[string]attr.Type{
											"nested_test": types.Float64Type,
										},
										map[string]attr.Value{
											"nested_test": types.Float64Unknown(),
										},
									),
								},
							),
						).AtName("nested_test"),
					),
				},
				PlanValue: types.Float64Unknown(),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Float64Response{
				PlanValue: testCase.request.PlanValue,
			}

			float64planmodifier.UseStateForUnknown().PlanModifyFloat64(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
