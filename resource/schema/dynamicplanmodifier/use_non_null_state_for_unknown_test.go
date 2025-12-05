// Copyright IBM Corp. 2021, 2025
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

func TestUseNonNullStateForUnknownModifierPlanModifyDynamic(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.DynamicRequest
		expected *planmodifier.DynamicResponse
	}{
		"null-state": {
			// when we first create the resource, the state value will be null,
			// so use the unknown value
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
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicUnknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.DynamicRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.DynamicPseudoType,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.String, "other"),
						},
					),
				},
				StateValue:  types.DynamicValue(types.StringValue("other")),
				PlanValue:   types.DynamicValue(types.StringValue("test")),
				ConfigValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringValue("test")),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state in
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
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicValue(types.StringValue("test")),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values should not be preserved
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
			expected: &planmodifier.DynamicResponse{
				PlanValue: types.DynamicUnknown(),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
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

			dynamicplanmodifier.UseNonNullStateForUnknown().PlanModifyDynamic(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
