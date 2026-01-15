// Copyright IBM Corp. 2021, 2025
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

func TestUseNonNullStateForUnknownModifierPlanModifyList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.ListRequest
		expected *planmodifier.ListResponse
	}{
		"null-state": {
			// when we first create the resource, the state value will be null,
			// so use the unknown value
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
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.List{ElementType: tftypes.String},
								[]tftypes.Value{
									tftypes.NewValue(tftypes.String, "other"),
								},
							),
						},
					),
				},
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				PlanValue:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state in
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.List{ElementType: tftypes.String},
								[]tftypes.Value{
									tftypes.NewValue(tftypes.String, "test"),
								},
							),
						},
					),
				},
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values should not be preserved
			request: planmodifier.ListRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.List{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.List{ElementType: tftypes.String},
								nil,
							),
						},
					),
				},
				StateValue:  types.ListNull(types.StringType),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListNull(types.StringType),
			},
			expected: &planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.ListRequest{
				StateValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				PlanValue:   types.ListUnknown(types.StringType),
				ConfigValue: types.ListUnknown(types.StringType),
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

			listplanmodifier.UseNonNullStateForUnknown().PlanModifyList(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
