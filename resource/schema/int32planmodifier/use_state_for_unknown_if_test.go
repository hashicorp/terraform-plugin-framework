// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int32planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Int32Request
		ifFunc   int32planmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.Int32Response
	}{
		"null-state": {
			request: planmodifier.Int32Request{
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
				StateValue:  types.Int32Null(),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown(),
			},
		},
		"known-plan": {
			request: planmodifier.Int32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 123),
						},
					),
				},
				StateValue:  types.Int32Value(123),
				PlanValue:   types.Int32Value(456),
				ConfigValue: types.Int32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Value(456),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.Int32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 123),
						},
					),
				},
				StateValue:  types.Int32Value(123),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Value(123),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.Int32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 123),
						},
					),
				},
				StateValue:  types.Int32Value(123),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown(),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.Int32Request{
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
				StateValue:  types.Int32Null(),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Null(),
			},
		},
		"unknown-config": {
			request: planmodifier.Int32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 123),
						},
					),
				},
				StateValue:  types.Int32Value(123),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Unknown(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Int32Response{
				PlanValue: testCase.request.PlanValue,
			}

			int32planmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyInt32(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
