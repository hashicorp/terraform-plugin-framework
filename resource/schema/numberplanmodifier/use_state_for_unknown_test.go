// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package numberplanmodifier_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUseStateForUnknownModifierPlanModifyNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.NumberRequest
		expected *planmodifier.NumberResponse
	}{
		"null-state": {
			// when we first create the resource, use the unknown
			// value
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberValue(big.NewFloat(2.4)),
				PlanValue:   types.NumberValue(big.NewFloat(1.2)),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberValue(big.NewFloat(1.2)),
			},
		},
		"non-null-state-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberValue(big.NewFloat(1.2)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberValue(big.NewFloat(1.2)),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.NumberRequest{
				StateValue:  types.NumberValue(big.NewFloat(1.2)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberUnknown(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
		"under-list": {
			request: planmodifier.NumberRequest{
				ConfigValue: types.NumberNull(),
				Path:        path.Root("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:   types.NumberUnknown(),
				StateValue:  types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
					),
				},
				PlanValue: types.NumberUnknown(),
			},
		},
		"under-set": {
			request: planmodifier.NumberRequest{
				ConfigValue: types.NumberNull(),
				Path: path.Root("test").AtSetValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_test": types.NumberType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_test": types.NumberType,
								},
								map[string]attr.Value{
									"nested_test": types.NumberUnknown(),
								},
							),
						},
					),
				).AtName("nested_test"),
				PlanValue:  types.NumberUnknown(),
				StateValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.SetValueMust(
								types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_test": types.NumberType,
									},
								},
								[]attr.Value{
									types.ObjectValueMust(
										map[string]attr.Type{
											"nested_test": types.NumberType,
										},
										map[string]attr.Value{
											"nested_test": types.NumberUnknown(),
										},
									),
								},
							),
						).AtName("nested_test"),
					),
				},
				PlanValue: types.NumberUnknown(),
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

			numberplanmodifier.UseStateForUnknown().PlanModifyNumber(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
