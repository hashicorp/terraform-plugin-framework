// Copyright IBM Corp. 2021, 2025
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

func TestUseNonNullStateForUnknownModifierPlanModifyObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.ObjectRequest
		expected *planmodifier.ObjectResponse
	}{
		"null-state": {
			// when we first create the resource, the state value will be null,
			// so use the unknown value
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
				StateValue:  types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
				PlanValue:   types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
				ConfigValue: types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
			},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
								map[string]tftypes.Value{
									"testattr": tftypes.NewValue(tftypes.String, "other"),
								},
							),
						},
					),
				},
				StateValue:  types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("other")}),
				PlanValue:   types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
				ConfigValue: types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
			},
		},
		"non-null-state-value-unknown-plan": {
			// this is the situation we want to preserve the state in
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
								map[string]tftypes.Value{
									"testattr": tftypes.NewValue(tftypes.String, "test"),
								},
							),
						},
					),
				},
				StateValue:  types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
				PlanValue:   types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
				ConfigValue: types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
			},
		},
		"null-state-value-unknown-plan": {
			// Null state values should not be preserved
			request: planmodifier.ObjectRequest{
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"attr": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
							},
						},
						map[string]tftypes.Value{
							"attr": tftypes.NewValue(
								tftypes.Object{AttributeTypes: map[string]tftypes.Type{"testattr": tftypes.String}},
								nil,
							),
						},
					),
				},
				StateValue:  types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
				PlanValue:   types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
				ConfigValue: types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.ObjectRequest{
				StateValue:  types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
				PlanValue:   types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
				ConfigValue: types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue: types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.ObjectResponse{
				PlanValue: testCase.request.PlanValue,
			}

			objectplanmodifier.UseNonNullStateForUnknown().PlanModifyObject(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
