// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamicplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceIfConfiguredModifierPlanModifyDynamic(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.DynamicAttribute{},
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

	testPlan := func(value types.Dynamic) tfsdk.Plan {
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

	testState := func(value types.Dynamic) tfsdk.State {
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
		request  planmodifier.DynamicRequest
		expected *planmodifier.DynamicResponse
	}{
		"state-null": {
			// resource creation
			request: planmodifier.DynamicRequest{
				ConfigValue: types.DynamicValue(types.StringValue("test")),
				Plan:        testPlan(types.DynamicValue(types.StringValue("test"))),
				PlanValue:   types.DynamicValue(types.StringValue("test")),
				State:       nullState,
				StateValue:  types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("test")),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.DynamicRequest{
				ConfigValue: types.DynamicNull(),
				Plan:        nullPlan,
				PlanValue:   types.DynamicNull(),
				State:       testState(types.DynamicValue(types.StringValue("test"))),
				StateValue:  types.DynamicValue(types.StringValue("test")),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicNull(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-configured": {
			request: planmodifier.DynamicRequest{
				ConfigValue: types.DynamicValue(types.StringValue("other")),
				Plan:        testPlan(types.DynamicValue(types.StringValue("other"))),
				PlanValue:   types.DynamicValue(types.StringValue("other")),
				State:       testState(types.DynamicValue(types.StringValue("test"))),
				StateValue:  types.DynamicValue(types.StringValue("test")),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("other")),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-different-unconfigured": {
			request: planmodifier.DynamicRequest{
				ConfigValue: types.DynamicNull(),
				Plan:        testPlan(types.DynamicValue(types.StringValue("other"))),
				PlanValue:   types.DynamicValue(types.StringValue("other")),
				State:       testState(types.DynamicValue(types.StringValue("test"))),
				StateValue:  types.DynamicValue(types.StringValue("test")),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("other")),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-unconfigured-underlying-value": {
			request: planmodifier.DynamicRequest{
				ConfigValue: types.DynamicValue(types.StringNull()),
				Plan:        testPlan(types.DynamicValue(types.StringValue("other"))),
				PlanValue:   types.DynamicValue(types.StringValue("other")),
				State:       testState(types.DynamicValue(types.StringValue("test"))),
				StateValue:  types.DynamicValue(types.StringValue("test")),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("other")),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.DynamicRequest{
				ConfigValue: types.DynamicValue(types.StringValue("test")),
				Plan:        testPlan(types.DynamicValue(types.StringValue("test"))),
				PlanValue:   types.DynamicValue(types.StringValue("test")),
				State:       testState(types.DynamicValue(types.StringValue("test"))),
				StateValue:  types.DynamicValue(types.StringValue("test")),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("test")),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.DynamicResponse{
				PlanValue: testCase.request.PlanValue,
			}

			dynamicplanmodifier.RequiresReplaceIfConfigured().PlanModifyDynamic(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
