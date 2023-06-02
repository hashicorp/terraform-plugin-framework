// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package objectplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceModifierPlanModifyObject(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{"testattr": types.StringType},
			},
		},
	}

	nullPlan := tfsdk.Plan{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			testSchema.Type().TerraformType(context.Background()),
			nil,
		),
	}

	nullState := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			testSchema.Type().TerraformType(context.Background()),
			nil,
		),
	}

	testPlan := func(value types.Object) tfsdk.Plan {
		tfValue, err := value.ToTerraformValue(context.Background())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.Plan{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(context.Background()),
				map[string]tftypes.Value{
					"testattr": tfValue,
				},
			),
		}
	}

	testState := func(value types.Object) tfsdk.State {
		tfValue, err := value.ToTerraformValue(context.Background())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.State{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(context.Background()),
				map[string]tftypes.Value{
					"testattr": tfValue,
				},
			),
		}
	}

	testCases := map[string]struct {
		request  planmodifier.ObjectRequest
		expected *planmodifier.ObjectResponse
	}{
		"state-null": {
			// resource creation
			request: planmodifier.ObjectRequest{
				Plan:       testPlan(types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType})),
				PlanValue:  types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
				State:      nullState,
				StateValue: types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue:       types.ObjectUnknown(map[string]attr.Type{"testattr": types.StringType}),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.ObjectRequest{
				Plan:       nullPlan,
				PlanValue:  types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
				State:      testState(types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")})),
				StateValue: types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue:       types.ObjectNull(map[string]attr.Type{"testattr": types.StringType}),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different": {
			request: planmodifier.ObjectRequest{
				Plan:       testPlan(types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("other")})),
				PlanValue:  types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("other")}),
				State:      testState(types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")})),
				StateValue: types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue:       types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("other")}),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.ObjectRequest{
				Plan:       testPlan(types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")})),
				PlanValue:  types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
				State:      testState(types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")})),
				StateValue: types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
			},
			expected: &planmodifier.ObjectResponse{
				PlanValue:       types.ObjectValueMust(map[string]attr.Type{"testattr": types.StringType}, map[string]attr.Value{"testattr": types.StringValue("test")}),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.ObjectResponse{
				PlanValue: testCase.request.PlanValue,
			}

			objectplanmodifier.RequiresReplace().PlanModifyObject(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
