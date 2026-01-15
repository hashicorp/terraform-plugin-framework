// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package boolplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownModifierPlanModifyBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.BoolRequest
		expected *planmodifier.BoolResponse
	}{
		"null-state": {
			// when we first create the resource, use the unknown
			// value
			request: planmodifier.BoolRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Bool,
							},
						},
						nil,
					),
				},
				StateValue:  types.BoolNull(),
				PlanValue:   types.BoolUnknown(),
				ConfigValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolUnknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.BoolRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Bool,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Bool, false),
						},
					),
				},
				StateValue:  types.BoolValue(false),
				PlanValue:   types.BoolValue(true),
				ConfigValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolValue(true),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			request: planmodifier.BoolRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Bool,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Bool, true),
						},
					),
				},
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolUnknown(),
				ConfigValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolValue(true),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values are still known, so we should preserve this as well.
			request: planmodifier.BoolRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Bool,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Bool, nil),
						},
					),
				},
				StateValue:  types.BoolNull(),
				PlanValue:   types.BoolUnknown(),
				ConfigValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolNull(),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.BoolRequest{
				StateValue:  types.BoolValue(true),
				PlanValue:   types.BoolUnknown(),
				ConfigValue: types.BoolUnknown(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue: types.BoolUnknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.BoolResponse{
				PlanValue: testCase.request.PlanValue,
			}

			boolplanmodifier.UseStateForUnknown().PlanModifyBool(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
