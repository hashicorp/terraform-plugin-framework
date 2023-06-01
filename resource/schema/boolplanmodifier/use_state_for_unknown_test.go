// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package boolplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUseStateForUnknownModifierPlanModifyBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.BoolRequest
		expected *planmodifier.BoolResponse
	}{
		"null-state": {
			// when we first create the resource, use the unknown
			// value
			request: planmodifier.BoolRequest{
				StateValue:  types.BoolNull(),
				PlanValue:   types.BoolUnknown(),
				ConfigValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolUnknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(false),
				PlanValue:   types.BoolValue(true),
				ConfigValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolValue(true),
			},
		},
		"non-null-state-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			request: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolUnknown(),
				ConfigValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolValue(true),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolUnknown(),
				ConfigValue: types.BoolUnknown(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolUnknown(),
			},
		},
		"under-list": {
			request: planmodifier.BoolRequest{
				ConfigValue: types.BoolNull(),
				Path:        path.Root("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:   types.BoolUnknown(),
				StateValue:  types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
					),
				},
				PlanValue: types.BoolUnknown(),
			},
		},
		"under-set": {
			request: planmodifier.BoolRequest{
				ConfigValue: types.BoolNull(),
				Path: path.Root("test").AtSetValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_test": types.BoolType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_test": types.BoolType,
								},
								map[string]attr.Value{
									"nested_test": types.BoolUnknown(),
								},
							),
						},
					),
				).AtName("nested_test"),
				PlanValue:  types.BoolUnknown(),
				StateValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.SetValueMust(
								types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_test": types.BoolType,
									},
								},
								[]attr.Value{
									types.ObjectValueMust(
										map[string]attr.Type{
											"nested_test": types.BoolType,
										},
										map[string]attr.Value{
											"nested_test": types.BoolUnknown(),
										},
									),
								},
							),
						).AtName("nested_test"),
					),
				},
				PlanValue: types.BoolUnknown(),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.BoolResponse{
				PlanValue: testCase.request.PlanValue,
			}

			boolplanmodifier.UseStateForUnknown().PlanModifyBool(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
