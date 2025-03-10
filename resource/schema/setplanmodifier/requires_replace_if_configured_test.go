// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package setplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceIfConfiguredModifierPlanModifySet(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.SetAttribute{
				ElementType: types.StringType,
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

	testPlan := func(value types.Set) tfsdk.Plan {
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

	testState := func(value types.Set) tfsdk.State {
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
		request  planmodifier.SetRequest
		expected *planmodifier.SetResponse
	}{
		"state-null": {
			// resource creation
			request: planmodifier.SetRequest{
				ConfigValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				Plan:        testPlan(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				PlanValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				State:       nullState,
				StateValue:  types.SetNull(types.StringType),
			},
			expected: &planmodifier.SetResponse{
				PlanValue:       types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.SetRequest{
				ConfigValue: types.SetNull(types.StringType),
				Plan:        nullPlan,
				PlanValue:   types.SetNull(types.StringType),
				State:       testState(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			expected: &planmodifier.SetResponse{
				PlanValue:       types.SetNull(types.StringType),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-configured": {
			request: planmodifier.SetRequest{
				ConfigValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				Plan:        testPlan(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other")})),
				PlanValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				State:       testState(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			expected: &planmodifier.SetResponse{
				PlanValue:       types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-different-unconfigured": {
			request: planmodifier.SetRequest{
				ConfigValue: types.SetNull(types.StringType),
				Plan:        testPlan(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other")})),
				PlanValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				State:       testState(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			expected: &planmodifier.SetResponse{
				PlanValue:       types.SetValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.SetRequest{
				ConfigValue: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				Plan:        testPlan(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				PlanValue:   types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				State:       testState(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				StateValue:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			expected: &planmodifier.SetResponse{
				PlanValue:       types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.SetResponse{
				PlanValue: testCase.request.PlanValue,
			}

			setplanmodifier.RequiresReplaceIfConfigured().PlanModifySet(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
