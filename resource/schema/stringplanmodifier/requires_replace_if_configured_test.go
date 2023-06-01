// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stringplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceIfConfiguredModifierPlanModifyString(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.StringAttribute{},
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

	testPlan := func(value types.String) tfsdk.Plan {
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

	testState := func(value types.String) tfsdk.State {
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
		request  planmodifier.StringRequest
		expected *planmodifier.StringResponse
	}{
		"state-null": {
			// resource creation
			request: planmodifier.StringRequest{
				ConfigValue: types.StringValue("test"),
				Plan:        testPlan(types.StringValue("test")),
				PlanValue:   types.StringValue("test"),
				State:       nullState,
				StateValue:  types.StringNull(),
			},
			expected: &planmodifier.StringResponse{
				PlanValue:       types.StringValue("test"),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.StringRequest{
				ConfigValue: types.StringNull(),
				Plan:        nullPlan,
				PlanValue:   types.StringNull(),
				State:       testState(types.StringValue("test")),
				StateValue:  types.StringValue("test"),
			},
			expected: &planmodifier.StringResponse{
				PlanValue:       types.StringNull(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-configured": {
			request: planmodifier.StringRequest{
				ConfigValue: types.StringValue("other"),
				Plan:        testPlan(types.StringValue("other")),
				PlanValue:   types.StringValue("other"),
				State:       testState(types.StringValue("test")),
				StateValue:  types.StringValue("test"),
			},
			expected: &planmodifier.StringResponse{
				PlanValue:       types.StringValue("other"),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-different-unconfigured": {
			request: planmodifier.StringRequest{
				ConfigValue: types.StringNull(),
				Plan:        testPlan(types.StringValue("other")),
				PlanValue:   types.StringValue("other"),
				State:       testState(types.StringValue("test")),
				StateValue:  types.StringValue("test"),
			},
			expected: &planmodifier.StringResponse{
				PlanValue:       types.StringValue("other"),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.StringRequest{
				ConfigValue: types.StringValue("test"),
				Plan:        testPlan(types.StringValue("test")),
				PlanValue:   types.StringValue("test"),
				State:       testState(types.StringValue("test")),
				StateValue:  types.StringValue("test"),
			},
			expected: &planmodifier.StringResponse{
				PlanValue:       types.StringValue("test"),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.StringResponse{
				PlanValue: testCase.request.PlanValue,
			}

			stringplanmodifier.RequiresReplaceIfConfigured().PlanModifyString(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
