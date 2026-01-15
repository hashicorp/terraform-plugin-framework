// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package int32planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseNonNullStateForUnknownModifierPlanModifyInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Int32Request
		expected *planmodifier.Int32Response
	}{
		"null-state": {
			// when we first create the resource, the state value will be null,
			// so use the unknown value
			request: planmodifier.Int32Request{
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
				StateValue:  types.Int32Null(),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.Int32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 1),
						},
					),
				},
				StateValue:  types.Int32Value(1),
				PlanValue:   types.Int32Value(2),
				ConfigValue: types.Int32Null(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Value(2),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state in
			request: planmodifier.Int32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 2),
						},
					),
				},
				StateValue:  types.Int32Value(2),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Value(2),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values should not be preserved
			request: planmodifier.Int32Request{
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
				StateValue:  types.Int32Null(),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Null(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown(),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.Int32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 2),
						},
					),
				},
				StateValue:  types.Int32Value(2),
				PlanValue:   types.Int32Unknown(),
				ConfigValue: types.Int32Unknown(),
			},
			expected: &planmodifier.Int32Response{
				PlanValue: types.Int32Unknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Int32Response{
				PlanValue: testCase.request.PlanValue,
			}

			int32planmodifier.UseNonNullStateForUnknown().PlanModifyInt32(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
