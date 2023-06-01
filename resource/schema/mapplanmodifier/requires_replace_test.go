// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiresReplaceModifierPlanModifyMap(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.MapAttribute{
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

	testPlan := func(value types.Map) tfsdk.Plan {
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

	testState := func(value types.Map) tfsdk.State {
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
		request  planmodifier.MapRequest
		expected *planmodifier.MapResponse
	}{
		"state-null": {
			// resource creation
			request: planmodifier.MapRequest{
				Plan:       testPlan(types.MapUnknown(types.StringType)),
				PlanValue:  types.MapUnknown(types.StringType),
				State:      nullState,
				StateValue: types.MapNull(types.StringType),
			},
			expected: &planmodifier.MapResponse{
				PlanValue:       types.MapUnknown(types.StringType),
				RequiresReplace: false,
			},
		},
		"plan-null": {
			// resource destroy
			request: planmodifier.MapRequest{
				Plan:       nullPlan,
				PlanValue:  types.MapNull(types.StringType),
				State:      testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")})),
				StateValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
			},
			expected: &planmodifier.MapResponse{
				PlanValue:       types.MapNull(types.StringType),
				RequiresReplace: false,
			},
		},
		"planvalue-statevalue-different": {
			request: planmodifier.MapRequest{
				Plan:       testPlan(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("other")})),
				PlanValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("other")}),
				State:      testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")})),
				StateValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
			},
			expected: &planmodifier.MapResponse{
				PlanValue:       types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("other")}),
				RequiresReplace: true,
			},
		},
		"planvalue-statevalue-equal": {
			request: planmodifier.MapRequest{
				Plan:       testPlan(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")})),
				PlanValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				State:      testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")})),
				StateValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
			},
			expected: &planmodifier.MapResponse{
				PlanValue:       types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				RequiresReplace: false,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &planmodifier.MapResponse{
				PlanValue: testCase.request.PlanValue,
			}

			mapplanmodifier.RequiresReplace().PlanModifyMap(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
