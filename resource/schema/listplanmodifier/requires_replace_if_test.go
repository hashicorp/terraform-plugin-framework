// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceIfModifierPlanModifyList(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.ListAttribute{
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

	testPlan := func(value types.List) tfsdk.Plan {
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

	testState := func(value types.List) tfsdk.State {
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
		request  planmodifier.ListRequest
		ifFunc   listplanmodifier.RequiresReplaceIfFunc
		expected *planmodifier.ListResponse
	}{
		"state-null": {
			// resource creation
			request: planmodifier.ListRequest{
				Plan:       testPlan(types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.ListResponse{
				PlanValue:       types.ListUnknown(types.StringType),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.ListRequest{
				Plan:       nullPlan,
				PlanValue:  types.ListNull(types.StringType),
				State:      testState(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				StateValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.ListResponse{
				PlanValue:       types.ListNull(types.StringType),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-if-false": {
			request: planmodifier.ListRequest{
				Plan:       testPlan(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other")})),
				PlanValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				State:      testState(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				StateValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = false // no change
			},
			expected: &planmodifier.ListResponse{
				PlanValue:       types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different-if-true": {
			request: planmodifier.ListRequest{
				Plan:       testPlan(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other")})),
				PlanValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				State:      testState(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				StateValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should reach here
			},
			expected: &planmodifier.ListResponse{
				PlanValue:       types.ListValueMust(types.StringType, []attr.Value{types.StringValue("other")}),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.ListRequest{
				Plan:       testPlan(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				PlanValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				State:      testState(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")})),
				StateValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			ifFunc: func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = true // should never reach here
			},
			expected: &planmodifier.ListResponse{
				PlanValue:       types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.ListResponse{
				PlanValue: testCase.request.PlanValue,
			}

			listplanmodifier.RequiresReplaceIf(testCase.ifFunc, "test", "test").PlanModifyList(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
