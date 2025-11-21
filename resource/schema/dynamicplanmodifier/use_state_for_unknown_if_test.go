// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamicplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyDynamic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.DynamicRequest
		ifFunc   dynamicplanmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.DynamicResponse
	}{
		"null-state": {
			request: planmodifier.DynamicRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.DynamicPseudoType,
							},
						},
						nil,
					),
				},
				StateValue:  types.DynamicNull(),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicUnknown(),
			},
		},
		"known-plan": {
			request: planmodifier.DynamicRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.DynamicPseudoType,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.String, "test"),
						},
					),
				},
				StateValue:  types.DynamicValue(types.StringValue("test")),
				PlanValue:   types.DynamicValue(types.StringValue("updated")),
				ConfigValue: types.DynamicNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringValue("updated")),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.DynamicRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.DynamicPseudoType,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.String, "test"),
						},
					),
				},
				StateValue:  types.DynamicValue(types.StringValue("test")),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringValue("test")),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.DynamicRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.DynamicPseudoType,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.String, "test"),
						},
					),
				},
				StateValue:  types.DynamicValue(types.StringValue("test")),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicUnknown(),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.DynamicRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.DynamicPseudoType,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
						},
					),
				},
				StateValue:  types.DynamicNull(),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicNull(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicNull(),
			},
		},
		"unknown-config": {
			request: planmodifier.DynamicRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.DynamicPseudoType,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.String, "test"),
						},
					),
				},
				StateValue:  types.DynamicValue(types.StringValue("test")),
				PlanValue:   types.DynamicUnknown(),
				ConfigValue: types.DynamicUnknown(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicUnknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.DynamicResponse{
				PlanValue: testCase.request.PlanValue,
			}

			dynamicplanmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyDynamic(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
