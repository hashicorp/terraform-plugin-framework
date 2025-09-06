// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyMap(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.MapRequest
		ifFunc   mapplanmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.MapResponse
	}{
		"null-state": {
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						nil,
					),
				},
				StateValue:  types.MapNull(types.StringType),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapUnknown(types.StringType),
			},
		},
		"known-plan": {
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
								"key1": tftypes.NewValue(tftypes.String, "value1"),
								"key2": tftypes.NewValue(tftypes.String, "value2"),
							}),
						},
					),
				},
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("value1"), "key2": types.StringValue("value2")}),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"key3": types.StringValue("value3"), "key4": types.StringValue("value4")}),
				ConfigValue: types.MapNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapValueMust(types.StringType, map[string]attr.Value{"key3": types.StringValue("value3"), "key4": types.StringValue("value4")}),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
								"key1": tftypes.NewValue(tftypes.String, "value1"),
								"key2": tftypes.NewValue(tftypes.String, "value2"),
							}),
						},
					),
				},
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("value1"), "key2": types.StringValue("value2")}),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("value1"), "key2": types.StringValue("value2")}),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
								"key1": tftypes.NewValue(tftypes.String, "value1"),
								"key2": tftypes.NewValue(tftypes.String, "value2"),
							}),
						},
					),
				},
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("value1"), "key2": types.StringValue("value2")}),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapUnknown(types.StringType),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
						},
					),
				},
				StateValue:  types.MapNull(types.StringType),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapNull(types.StringType),
			},
		},
		"unknown-config": {
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
								"key1": tftypes.NewValue(tftypes.String, "value1"),
								"key2": tftypes.NewValue(tftypes.String, "value2"),
							}),
						},
					),
				},
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("value1"), "key2": types.StringValue("value2")}),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapUnknown(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapUnknown(types.StringType),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.MapResponse{
				PlanValue: testCase.request.PlanValue,
			}

			mapplanmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyMap(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
