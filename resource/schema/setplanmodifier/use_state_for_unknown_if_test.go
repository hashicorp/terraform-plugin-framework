// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package setplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifySet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.SetRequest
		ifFunc   setplanmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.SetResponse
	}{
		"null-state": {
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						nil,
					),
				},
				StateValue:  types.SetNull(types.StringType),
				PlanValue:   types.SetUnknown(types.StringType),
				ConfigValue: types.SetNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetUnknown(types.StringType),
			},
		},
		"known-plan": {
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							}),
						},
					),
				},
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
				PlanValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other1"), types.StringValue("other2")}),
				ConfigValue: types.SetNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other1"), types.StringValue("other2")}),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							}),
						},
					),
				},
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
				PlanValue:   types.SetUnknown(types.StringType),
				ConfigValue: types.SetNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							}),
						},
					),
				},
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
				PlanValue:   types.SetUnknown(types.StringType),
				ConfigValue: types.SetNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetUnknown(types.StringType),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
						},
					),
				},
				StateValue:  types.SetNull(types.StringType),
				PlanValue:   types.SetUnknown(types.StringType),
				ConfigValue: types.SetNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetNull(types.StringType),
			},
		},
		"unknown-config": {
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							}),
						},
					),
				},
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
				PlanValue:   types.SetUnknown(types.StringType),
				ConfigValue: types.SetUnknown(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetUnknown(types.StringType),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.SetResponse{
				PlanValue: testCase.request.PlanValue,
			}

			setplanmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifySet(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
