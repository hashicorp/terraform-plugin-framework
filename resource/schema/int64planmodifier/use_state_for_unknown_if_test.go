// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int64planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Int64Request
		ifFunc   int64planmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.Int64Response
	}{
		"null-state": {
			request: planmodifier.Int64Request{
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
				StateValue:  types.Int64Null(),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown(),
			},
		},
		"known-plan": {
			request: planmodifier.Int64Request{
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
				StateValue:  types.Int64Value(123),
				PlanValue:   types.Int64Value(456),
				ConfigValue: types.Int64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Value(456),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.Int64Request{
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
				StateValue:  types.Int64Value(123),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Value(123),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.Int64Request{
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
				StateValue:  types.Int64Value(123),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown(),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.Int64Request{
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
				StateValue:  types.Int64Null(),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Null(),
			},
		},
		"unknown-config": {
			request: planmodifier.Int64Request{
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
				StateValue:  types.Int64Value(123),
				PlanValue:   types.Int64Unknown(),
				ConfigValue: types.Int64Unknown(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.Int64Response{
				PlanValue: types.Int64Unknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Int64Response{
				PlanValue: testCase.request.PlanValue,
			}

			int64planmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyInt64(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
