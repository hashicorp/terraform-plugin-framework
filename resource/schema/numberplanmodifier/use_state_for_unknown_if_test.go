// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package numberplanmodifier_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.NumberRequest
		ifFunc   numberplanmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.NumberResponse
	}{
		"null-state": {
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						nil,
					),
				},
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.NumberRequest, resp *numberplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
		"known-plan": {
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 123.45),
						},
					),
				},
				StateValue:  types.NumberValue(big.NewFloat(123.45)),
				PlanValue:   types.NumberValue(big.NewFloat(456.78)),
				ConfigValue: types.NumberNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.NumberRequest, resp *numberplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberValue(big.NewFloat(456.78)),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 123.45),
						},
					),
				},
				StateValue:  types.NumberValue(big.NewFloat(123.45)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.NumberRequest, resp *numberplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberValue(big.NewFloat(123.45)),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 123.45),
						},
					),
				},
				StateValue:  types.NumberValue(big.NewFloat(123.45)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.NumberRequest, resp *numberplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, nil),
						},
					),
				},
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.NumberRequest, resp *numberplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberNull(),
			},
		},
		"unknown-config": {
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 123.45),
						},
					),
				},
				StateValue:  types.NumberValue(big.NewFloat(123.45)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberUnknown(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.NumberRequest, resp *numberplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.NumberResponse{
				PlanValue: testCase.request.PlanValue,
			}

			numberplanmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyNumber(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
