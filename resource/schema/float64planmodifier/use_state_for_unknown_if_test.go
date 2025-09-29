// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float64planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Float64Request
		ifFunc   float64planmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.Float64Response
	}{
		"null-state": {
			request: planmodifier.Float64Request{
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
				StateValue:  types.Float64Null(),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float64Request, resp *float64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown(),
			},
		},
		"known-plan": {
			request: planmodifier.Float64Request{
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
				StateValue:  types.Float64Value(123.45),
				PlanValue:   types.Float64Value(456.78),
				ConfigValue: types.Float64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float64Request, resp *float64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Value(456.78),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.Float64Request{
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
				StateValue:  types.Float64Value(123.45),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float64Request, resp *float64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Value(123.45),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.Float64Request{
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
				StateValue:  types.Float64Value(123.45),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float64Request, resp *float64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown(),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.Float64Request{
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
				StateValue:  types.Float64Null(),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float64Request, resp *float64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Null(),
			},
		},
		"unknown-config": {
			request: planmodifier.Float64Request{
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
				StateValue:  types.Float64Value(123.45),
				PlanValue:   types.Float64Unknown(),
				ConfigValue: types.Float64Unknown(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float64Request, resp *float64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Float64Response{
				PlanValue: types.Float64Unknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Float64Response{
				PlanValue: testCase.request.PlanValue,
			}

			float64planmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyFloat64(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
