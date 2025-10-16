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

func TestRequiresReplaceIfModifierPlanModifyDynamic(t *testing.T) {
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
		ifFunc   dynamicplanmodifier.RequiresReplaceIfFunc
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
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
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
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicNull(),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-if-false": {
			request: planmodifier.DynamicRequest{
				Plan:       testPlan(types.DynamicValue(types.StringValue("other"))),
				PlanValue:  types.DynamicValue(types.StringValue("other")),
				State:      testState(types.DynamicValue(types.StringValue("test"))),
				StateValue: types.DynamicValue(types.StringValue("test")),
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = false // no change
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("other")),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-if-true": {
			request: planmodifier.DynamicRequest{
				Plan:       testPlan(types.DynamicValue(types.StringValue("other"))),
				PlanValue:  types.DynamicValue(types.StringValue("other")),
				State:      testState(types.DynamicValue(types.StringValue("test"))),
				StateValue: types.DynamicValue(types.StringValue("test")),
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should reach here
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
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicValue(types.StringValue("test")),
				RequiresReplace: false,
			},
		},
		"write-only-with-null-config-value": {
			request: planmodifier.DynamicRequest{
				Plan:        testPlan(types.DynamicValue(types.StringValue("test"))),
				PlanValue:   types.DynamicNull(),
				State:       testState(types.DynamicValue(types.StringValue("test"))),
				StateValue:  types.DynamicNull(),
				ConfigValue: types.DynamicNull(),
				WriteOnly:   true,
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicNull(),
				RequiresReplace: false,
			},
		},
		"write-only-with-actual-config-value": {
			request: planmodifier.DynamicRequest{
				Plan:        testPlan(types.DynamicValue(types.StringValue("test"))),
				PlanValue:   types.DynamicNull(),
				State:       testState(types.DynamicValue(types.StringValue("test"))),
				StateValue:  types.DynamicNull(),
				ConfigValue: types.DynamicValue(types.StringValue("test value from config")),
				WriteOnly:   true,
			},
			ifFunc: func(ctx context.Context, req planmodifier.DynamicRequest, resp *dynamicplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true
			},
			expected: &planmodifier.DynamicResponse{
				PlanValue:       types.DynamicNull(),
				RequiresReplace: true,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.DynamicResponse{
				PlanValue: testCase.request.PlanValue,
			}

			dynamicplanmodifier.RequiresReplaceIf(testCase.ifFunc, "test", "test").PlanModifyDynamic(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
