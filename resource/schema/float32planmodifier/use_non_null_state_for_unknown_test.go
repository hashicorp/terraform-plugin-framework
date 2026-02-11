// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package float32planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseNonNullStateForUnknownModifierPlanModifyFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.Float32Request
		expected *planmodifier.Float32Response
	}{
		"null-state": {
			// when we first create the resource, the state value will be null,
			// so use the unknown value
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Null(),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown(),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Value(1.2),
				PlanValue:   types.Float32Value(2.4),
				ConfigValue: types.Float32Null(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Value(2.4),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state in
			request: planmodifier.Float32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 2.4),
						},
					),
				},
				StateValue:  types.Float32Value(2.4),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Value(2.4),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values should not be preserved
			request: planmodifier.Float32Request{
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
				StateValue:  types.Float32Null(),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Null(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown(),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.Float32Request{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Number,
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(tftypes.Number, 2.4),
						},
					),
				},
				StateValue:  types.Float32Value(2.4),
				PlanValue:   types.Float32Unknown(),
				ConfigValue: types.Float32Unknown(),
			},
			expected: &planmodifier.Float32Response{
				PlanValue: types.Float32Unknown(),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Float32Response{
				PlanValue: testCase.request.PlanValue,
			}

			float32planmodifier.UseNonNullStateForUnknown().PlanModifyFloat32(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
