// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package objectplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownIfModifierPlanModifyObject(t *testing.T) {
	t.Parallel()

	objTypes := map[string]attr.Type{"testattr": types.StringType}

	testCases := map[string]struct {
		request  planmodifier.ObjectRequest
		ifFunc   objectplanmodifier.UseStateForUnknownIfFunc
		expected *planmodifier.ObjectResponse
	}{
		"null-state": {
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						nil,
					),
				},
				StateValue:  types.ObjectNull(objTypes),
				PlanValue:   types.ObjectUnknown(objTypes),
				ConfigValue: types.ObjectNull(objTypes),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(objTypes),
			},
		},
		"known-plan": {
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}}, map[string]tftypes.Value{
								"testattr": tftypes.NewValue(tftypes.String, "test"),
							}),
						},
					),
				},
				StateValue:  types.ObjectValueMust(objTypes, map[string]attr.Value{"testattr": types.StringValue("test")}),
				PlanValue:   types.ObjectValueMust(objTypes, map[string]attr.Value{"testattr": types.StringValue("other")}),
				ConfigValue: types.ObjectNull(objTypes),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectValueMust(objTypes, map[string]attr.Value{"testattr": types.StringValue("other")}),
			},
		},
		"non-null-state-value-unknown-plan-if-true": {
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}}, map[string]tftypes.Value{
								"testattr": tftypes.NewValue(tftypes.String, "test"),
							}),
						},
					),
				},
				StateValue:  types.ObjectValueMust(objTypes, map[string]attr.Value{"testattr": types.StringValue("test")}),
				PlanValue:   types.ObjectUnknown(objTypes),
				ConfigValue: types.ObjectNull(objTypes),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectValueMust(objTypes, map[string]attr.Value{"testattr": types.StringValue("test")}),
			},
		},
		"non-null-state-value-unknown-plan-if-false": {
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}}, map[string]tftypes.Value{
								"testattr": tftypes.NewValue(tftypes.String, "test"),
							}),
						},
					),
				},
				StateValue:  types.ObjectValueMust(objTypes, map[string]attr.Value{"testattr": types.StringValue("test")}),
				PlanValue:   types.ObjectUnknown(objTypes),
				ConfigValue: types.ObjectNull(objTypes),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = false
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(objTypes),
			},
		},
		"null-state-value-unknown-plan-if-true": {
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}}, nil),
						},
					),
				},
				StateValue:  types.ObjectNull(objTypes),
				PlanValue:   types.ObjectUnknown(objTypes),
				ConfigValue: types.ObjectNull(objTypes),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectNull(objTypes),
			},
		},
		"unknown-config": {
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}}, map[string]tftypes.Value{
								"testattr": tftypes.NewValue(tftypes.String, "test"),
							}),
						},
					),
				},
				StateValue:  types.ObjectValueMust(objTypes, map[string]attr.Value{"testattr": types.StringValue("test")}),
				PlanValue:   types.ObjectUnknown(objTypes),
				ConfigValue: types.ObjectUnknown(objTypes),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.UseStateForUnknownIfFuncResponse) {
				resp.UseState = true // should never reach here
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(objTypes),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.ObjectResponse{
				PlanValue: testCase.request.PlanValue,
			}

			objectplanmodifier.UseStateForUnknownIf(testCase.ifFunc, "test description", "test markdown description").PlanModifyObject(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
