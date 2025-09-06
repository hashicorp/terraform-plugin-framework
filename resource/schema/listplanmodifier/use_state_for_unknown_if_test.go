// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.ListRequest
		ifFunc   listplanmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.ListResponse
	}{
		"null-state": {
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						nil,
					),
				},
				StateValue:  types.ListNull(types.StringType),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"known-plan": {
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							}),
						},
					),
				},
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
				PlanValue:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other1"), types.StringValue("other2")}),
				ConfigValue: types.ListNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other1"), types.StringValue("other2")}),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							}),
						},
					),
				},
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							}),
						},
					),
				},
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						},
					),
				},
				StateValue:  types.ListNull(types.StringType),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListNull(types.StringType),
			},
		},
		"unknown-config": {
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							}),
						},
					),
				},
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test1"), types.StringValue("test2")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListUnknown(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.ListResponse{
				PlanValue: testCase.request.PlanValue,
			}

			listplanmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyList(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
