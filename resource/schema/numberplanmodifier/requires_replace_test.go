// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package numberplanmodifier_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceModifierPlanModifyNumber(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.NumberAttribute{},
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

	testPlan := func(value types.Number) tfsdk.Plan {
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

	testState := func(value types.Number) tfsdk.State {
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
		request  planmodifier.NumberRequest
		expected *planmodifier.NumberResponse
	}{
		"state-null": {
			// resource creation
			request: planmodifier.NumberRequest{
				Plan:       testPlan(types.NumberUnknown()),
				PlanValue:  types.NumberUnknown(),
				State:      nullState,
				StateValue: types.NumberNull(),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue:       types.NumberUnknown(),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.NumberRequest{
				Plan:       nullPlan,
				PlanValue:  types.NumberNull(),
				State:      testState(types.NumberValue(big.NewFloat(1.2))),
				StateValue: types.NumberValue(big.NewFloat(1.2)),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue:       types.NumberNull(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different": {
			request: planmodifier.NumberRequest{
				Plan:       testPlan(types.NumberValue(big.NewFloat(2.4))),
				PlanValue:  types.NumberValue(big.NewFloat(2.4)),
				State:      testState(types.NumberValue(big.NewFloat(1.2))),
				StateValue: types.NumberValue(big.NewFloat(1.2)),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue:       types.NumberValue(big.NewFloat(2.4)),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.NumberRequest{
				Plan:       testPlan(types.NumberValue(big.NewFloat(1.2))),
				PlanValue:  types.NumberValue(big.NewFloat(1.2)),
				State:      testState(types.NumberValue(big.NewFloat(1.2))),
				StateValue: types.NumberValue(big.NewFloat(1.2)),
			},
			expected: &planmodifier.NumberResponse{
				PlanValue:       types.NumberValue(big.NewFloat(1.2)),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.NumberResponse{
				PlanValue: testCase.request.PlanValue,
			}

			numberplanmodifier.RequiresReplace().PlanModifyNumber(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
