// Copyright IBM Corp. 2021, 2025
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

func TestRequiresReplaceModifierPlanModifyDynamic(t *testing.T) {
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
				Plan:       testPlan(types.DynamicUnknown()),
				PlanValue:  types.DynamicUnknown(),
				State:      nullState,
				StateValue: types.DynamicNull(),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicUnknown(),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.DynamicRequest{
				Plan:       nullPlan,
				PlanValue:  types.DynamicNull(),
				State:      testState(types.DynamicValue(types.StringValue("test"))),
				StateValue: types.DynamicValue(types.StringValue("test")),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicNull(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different": {
			request: planmodifier.DynamicRequest{
				Plan:       testPlan(types.DynamicValue(types.StringValue("other"))),
				PlanValue:  types.DynamicValue(types.StringValue("other")),
				State:      testState(types.DynamicValue(types.StringValue("test"))),
				StateValue: types.DynamicValue(types.StringValue("test")),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("other")),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.DynamicRequest{
				Plan:       testPlan(types.DynamicValue(types.StringValue("test"))),
				PlanValue:  types.DynamicValue(types.StringValue("test")),
				State:      testState(types.DynamicValue(types.StringValue("test"))),
				StateValue: types.DynamicValue(types.StringValue("test")),
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("test")),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.DynamicResponse{
				PlanValue: testCase.request.PlanValue,
			}

			dynamicplanmodifier.RequiresReplace().PlanModifyDynamic(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
