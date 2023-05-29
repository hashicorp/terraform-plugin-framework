// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package boolplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceModifierPlanModifyBool(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.BoolAttribute{},
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

	testPlan := func(value types.Bool) tfsdk.Plan {
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

	testState := func(value types.Bool) tfsdk.State {
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
		request  planmodifier.BoolRequest
		expected *planmodifier.BoolResponse
	}{
		"state-null": {
			// resource creation
			request: planmodifier.BoolRequest{
				Plan:       testPlan(types.BoolUnknown()),
				PlanValue:  types.BoolUnknown(),
				State:      nullState,
				StateValue: types.BoolNull(),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue:       types.BoolUnknown(),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.BoolRequest{
				Plan:       nullPlan,
				PlanValue:  types.BoolNull(),
				State:      testState(types.BoolValue(true)),
				StateValue: types.BoolValue(true),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue:       types.BoolNull(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different": {
			request: planmodifier.BoolRequest{
				Plan:       testPlan(types.BoolValue(false)),
				PlanValue:  types.BoolValue(false),
				State:      testState(types.BoolValue(true)),
				StateValue: types.BoolValue(true),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue:       types.BoolValue(false),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.BoolRequest{
				Plan:       testPlan(types.BoolValue(true)),
				PlanValue:  types.BoolValue(true),
				State:      testState(types.BoolValue(true)),
				StateValue: types.BoolValue(true),
			},
			expected: &planmodifier.BoolResponse{
				PlanValue:       types.BoolValue(true),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.BoolResponse{
				PlanValue: testCase.request.PlanValue,
			}

			boolplanmodifier.RequiresReplace().PlanModifyBool(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
