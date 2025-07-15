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

func TestUseStateForUnknownModifierPlanModifyMap(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.MapRequest
		expected *planmodifier.MapResponse
	}{
		"null-state": {
			// when we first create the resource, use the unknown
			// value
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
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapUnknown(types.StringType),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Map{ElementType: tftypes.String},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(tftypes.String, "other"),
								},
							),
						},
					),
				},
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("other")}),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				ConfigValue: types.MapNull(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Map{ElementType: tftypes.String},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(tftypes.String, "test"),
								},
							),
						},
					),
				},
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapNull(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values are still known, so we should preserve this as well.
			request: planmodifier.MapRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Map{ElementType: tftypes.String},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Map{ElementType: tftypes.String},
								nil,
							),
						},
					),
				},
				StateValue:  types.MapNull(types.StringType),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapNull(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue: types.MapNull(types.StringType),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.MapRequest{
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				PlanValue:   types.MapUnknown(types.StringType),
				ConfigValue: types.MapUnknown(types.StringType),
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

			mapplanmodifier.UseStateForUnknown().PlanModifyMap(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
