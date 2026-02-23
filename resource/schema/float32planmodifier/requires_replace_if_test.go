// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package float32planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceIfModifierPlanModifyFloat32(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.Float32Attribute{},
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

	testPlan := func(value types.Float32) tfsdk.Plan {
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

	testState := func(value types.Float32) tfsdk.State {
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
		request  planmodifier.Float32Request
		ifFunc   float32planmodifier.RequiresReplaceIfFunc
		expected *planmodifier.Float32Response
	}{
		"state-null": {
			// resource creation
			request: planmodifier.Float32Request{
				Plan:       testPlan(types.Float32Unknown()),
				PlanValue:  types.Float32Unknown(),
				State:      nullState,
				StateValue: types.Float32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.Float32Response{
				PlanValue:       types.Float32Unknown(),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.Float32Request{
				Plan:       nullPlan,
				PlanValue:  types.Float32Null(),
				State:      testState(types.Float32Value(1.2)),
				StateValue: types.Float32Value(1.2),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.Float32Response{
				PlanValue:       types.Float32Null(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-if-false": {
			request: planmodifier.Float32Request{
				Plan:       testPlan(types.Float32Value(2.4)),
				PlanValue:  types.Float32Value(2.4),
				State:      testState(types.Float32Value(1.2)),
				StateValue: types.Float32Value(1.2),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = false // no change
			},
			expected: &planmodifier.Float32Response{
				PlanValue:       types.Float32Value(2.4),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-if-true": {
			request: planmodifier.Float32Request{
				Plan:       testPlan(types.Float32Value(2.4)),
				PlanValue:  types.Float32Value(2.4),
				State:      testState(types.Float32Value(1.2)),
				StateValue: types.Float32Value(1.2),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should reach here
			},
			expected: &planmodifier.Float32Response{
				PlanValue:       types.Float32Value(2.4),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.Float32Request{
				Plan:       testPlan(types.Float32Value(1.2)),
				PlanValue:  types.Float32Value(1.2),
				State:      testState(types.Float32Value(1.2)),
				StateValue: types.Float32Value(1.2),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Float32Request, resp *float32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.Float32Response{
				PlanValue:       types.Float32Value(1.2),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Float32Response{
				PlanValue: testCase.request.PlanValue,
			}

			float32planmodifier.RequiresReplaceIf(testCase.ifFunc, "test", "test").PlanModifyFloat32(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
