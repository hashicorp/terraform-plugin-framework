// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package int32planmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRequiresReplaceIfModifierPlanModifyInt32(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.Int32Attribute{},
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

	testPlan := func(value types.Int32) tfsdk.Plan {
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

	testState := func(value types.Int32) tfsdk.State {
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
		request  planmodifier.Int32Request
		ifFunc   int32planmodifier.RequiresReplaceIfFunc
		expected *planmodifier.Int32Response
	}{
		"state-null": {
			// resource creation
			request: planmodifier.Int32Request{
				Plan:       testPlan(types.Int32Unknown()),
				PlanValue:  types.Int32Unknown(),
				State:      nullState,
				StateValue: types.Int32Null(),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.Int32Response{
				PlanValue:       types.Int32Unknown(),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.Int32Request{
				Plan:       nullPlan,
				PlanValue:  types.Int32Null(),
				State:      testState(types.Int32Value(1)),
				StateValue: types.Int32Value(1),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.Int32Response{
				PlanValue:       types.Int32Null(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-if-false": {
			request: planmodifier.Int32Request{
				Plan:       testPlan(types.Int32Value(2)),
				PlanValue:  types.Int32Value(2),
				State:      testState(types.Int32Value(1)),
				StateValue: types.Int32Value(1),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = false // no change
			},
			expected: &planmodifier.Int32Response{
				PlanValue:       types.Int32Value(2),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-if-true": {
			request: planmodifier.Int32Request{
				Plan:       testPlan(types.Int32Value(2)),
				PlanValue:  types.Int32Value(2),
				State:      testState(types.Int32Value(1)),
				StateValue: types.Int32Value(1),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should reach here
			},
			expected: &planmodifier.Int32Response{
				PlanValue:       types.Int32Value(2),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.Int32Request{
				Plan:       testPlan(types.Int32Value(1)),
				PlanValue:  types.Int32Value(1),
				State:      testState(types.Int32Value(1)),
				StateValue: types.Int32Value(1),
			},
			ifFunc: func(ctx context.Context, req planmodifier.Int32Request, resp *int32planmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.Int32Response{
				PlanValue:       types.Int32Value(1),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.Int32Response{
				PlanValue: testCase.request.PlanValue,
			}

			int32planmodifier.RequiresReplaceIf(testCase.ifFunc, "test", "test").PlanModifyInt32(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
