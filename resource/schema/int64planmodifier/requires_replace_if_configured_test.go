// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int64planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceIfConfiguredModifierPlanModifyInt64(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.Int64Attribute{},
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

	testPlan := func(value types.Int64) tfsdk.Plan {
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

	testState := func(value types.Int64) tfsdk.State {
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
		request  planmodifier.Int64Request
		expected *planmodifier.Int64Response
	}{
		"state-null": {
			// resource creation
			request: planmodifier.Int64Request{
				ConfigValue: types.Int64Value(1),
				Plan:        testPlan(types.Int64Value(1)),
				PlanValue:   types.Int64Value(1),
				State:       nullState,
				StateValue:  types.Int64Null(),
			},
			expected: &planmodifier.Int64Response{
				PlanValue:       types.Int64Value(1),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.Int64Request{
				ConfigValue: types.Int64Null(),
				Plan:        nullPlan,
				PlanValue:   types.Int64Null(),
				State:       testState(types.Int64Value(1)),
				StateValue:  types.Int64Value(1),
			},
			expected: &planmodifier.Int64Response{
				PlanValue:       types.Int64Null(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-configured": {
			request: planmodifier.Int64Request{
				ConfigValue: types.Int64Value(2),
				Plan:        testPlan(types.Int64Value(2)),
				PlanValue:   types.Int64Value(2),
				State:       testState(types.Int64Value(1)),
				StateValue:  types.Int64Value(1),
			},
			expected: &planmodifier.Int64Response{
				PlanValue:       types.Int64Value(2),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-different-unconfigured": {
			request: planmodifier.Int64Request{
				ConfigValue: types.Int64Null(),
				Plan:        testPlan(types.Int64Value(2)),
				PlanValue:   types.Int64Value(2),
				State:       testState(types.Int64Value(1)),
				StateValue:  types.Int64Value(1),
			},
			expected: &planmodifier.Int64Response{
				PlanValue:       types.Int64Value(2),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.Int64Request{
				ConfigValue: types.Int64Value(1),
				Plan:        testPlan(types.Int64Value(1)),
				PlanValue:   types.Int64Value(1),
				State:       testState(types.Int64Value(1)),
				StateValue:  types.Int64Value(1),
			},
			expected: &planmodifier.Int64Response{
				PlanValue:       types.Int64Value(1),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Int64Response{
				PlanValue: testCase.request.PlanValue,
			}

			int64planmodifier.RequiresReplaceIfConfigured().PlanModifyInt64(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
