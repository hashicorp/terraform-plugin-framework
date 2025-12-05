// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package numberplanmodifier_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseNonNullStateForUnknownModifierPlanModifyNumber(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.NumberRequest
		expected *planmodifier.NumberResponse
	}{
		"null-state": {
			// when we first create the resource, the state value will be null,
			// so use the unknown value
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						nil,
					),
				},
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 1.2),
						},
					),
				},
				StateValue:  types.NumberValue(big.NewFloat(1.2)),
				PlanValue:   types.NumberValue(big.NewFloat(2.4)),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberValue(big.NewFloat(2.4)),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state in
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, big.NewFloat(2.4)),
						},
					),
				},
				StateValue:  types.NumberValue(big.NewFloat(2.4)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberValue(big.NewFloat(2.4)),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values should not be preserved
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, nil),
						},
					),
				},
				StateValue:  types.NumberNull(),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.NumberRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, big.NewFloat(2.4)),
						},
					),
				},
				StateValue:  types.NumberValue(big.NewFloat(2.4)),
				PlanValue:   types.NumberUnknown(),
				ConfigValue: types.NumberUnknown(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue: types.NumberUnknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.NumberResponse{
				PlanValue: testCase.request.PlanValue,
			}

			numberplanmodifier.UseNonNullStateForUnknown().PlanModifyNumber(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
