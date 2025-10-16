// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float32planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Float32Request
		ifFunc   float32planmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.Float32Response
	}{
		"null-state": {
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Null(),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown(),
			},
		},
		"known-plan": {
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Value(123.45),
				PlanValue:   types.Float32Value(456.78),
				ConfigValue: types.Float32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Value(456.78),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Value(123.45),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Value(123.45),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Value(123.45),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown(),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Null(),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Null(),
			},
		},
		"unknown-config": {
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Value(123.45),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Unknown(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Float32Response{
				PlanValue: testCase.request.PlanValue,
			}

			float32planmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyFloat32(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
