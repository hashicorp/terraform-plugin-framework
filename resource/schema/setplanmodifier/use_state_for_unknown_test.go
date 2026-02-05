// Copyright IBM Corp. 2021, 2026
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

func TestUseStateForUnknownModifierPlanModifySet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.SetRequest
		expected *planmodifier.SetResponse
	}{
		"null-state": {
			// when we first create the resource, use the unknown
			// value
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
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetUnknown(types.StringType),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Set{ElementType: tftypes.String},
								[]tftypes.Value{
									tftypes.NewValue(tftypes.String, "other"),
								},
							),
						},
					),
				},
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				PlanValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				ConfigValue: types.SetNull(types.StringType),
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Set{ElementType: tftypes.String},
								[]tftypes.Value{
									tftypes.NewValue(tftypes.String, "test"),
								},
							),
						},
					),
				},
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				PlanValue:   types.SetUnknown(types.StringType),
				ConfigValue: types.SetNull(types.StringType),
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values are still known, so we should preserve this as well.
			request: planmodifier.SetRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Set{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Set{ElementType: tftypes.String},
								nil,
							),
						},
					),
				},
				StateValue:  types.SetNull(types.StringType),
				PlanValue:   types.SetUnknown(types.StringType),
				ConfigValue: types.SetNull(types.StringType),
			},
			expected: &planmodifier.SetResponse{
				PlanValue: types.SetNull(types.StringType),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.SetRequest{
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				PlanValue:   types.SetUnknown(types.StringType),
				ConfigValue: types.SetUnknown(types.StringType),
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

			setplanmodifier.UseStateForUnknown().PlanModifySet(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
