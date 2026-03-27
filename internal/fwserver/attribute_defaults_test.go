// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testdefaults"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSetDefaultValueAtPath(t *testing.T) {
	t.Parallel()

	listDefaultValue := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("alpha"),
		types.StringValue("beta"),
	})

	objectDefaultValue := types.ObjectValueMust(
		map[string]attr.Type{
			"enabled": types.BoolType,
		},
		map[string]attr.Value{
			"enabled": types.BoolValue(true),
		},
	)

	testType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"alpha": tftypes.Bool,
		"beta":  tftypes.Number,
		"gamma": tftypes.List{ElementType: tftypes.String},
		"delta": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"enabled": tftypes.Bool,
		}},
		"epsilon": tftypes.DynamicPseudoType,
		"zeta":    tftypes.String,
	}}

	testSchema := testschema.Schema{Attributes: map[string]fwschema.Attribute{
		"alpha": testschema.AttributeWithBoolDefaultValue{
			Optional: true,
			Default: testdefaults.Bool{DefaultBoolMethod: func(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
				resp.PlanValue = types.BoolValue(true)
			}},
		},
		"beta": testschema.AttributeWithNumberDefaultValue{
			Optional: true,
			Default: testdefaults.Number{DefaultNumberMethod: func(_ context.Context, _ defaults.NumberRequest, resp *defaults.NumberResponse) {
				resp.PlanValue = types.NumberValue(big.NewFloat(42))
			}},
		},
		"gamma": testschema.AttributeWithListDefaultValue{
			Optional:    true,
			ElementType: types.StringType,
			Default: testdefaults.List{DefaultListMethod: func(_ context.Context, _ defaults.ListRequest, resp *defaults.ListResponse) {
				resp.PlanValue = listDefaultValue
			}},
		},
		"delta": testschema.AttributeWithObjectDefaultValue{
			Optional: true,
			AttributeTypes: map[string]attr.Type{
				"enabled": types.BoolType,
			},
			Default: testdefaults.Object{DefaultObjectMethod: func(_ context.Context, _ defaults.ObjectRequest, resp *defaults.ObjectResponse) {
				resp.PlanValue = objectDefaultValue
			}},
		},
		"epsilon": testschema.AttributeWithDynamicDefaultValue{
			Optional: true,
			Default: testdefaults.Dynamic{DefaultDynamicMethod: func(_ context.Context, _ defaults.DynamicRequest, resp *defaults.DynamicResponse) {
				resp.PlanValue = types.DynamicValue(types.StringValue("dynamic-default"))
			}},
		},
		"zeta": testschema.AttributeWithStringDefaultValue{
			Optional: true,
			Default: testdefaults.String{DefaultStringMethod: func(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
				resp.Diagnostics.AddError("default failed", "string default diagnostic")
			}},
		},
	}}

	baseConfig := tftypes.NewValue(testType, map[string]tftypes.Value{
		"alpha":   tftypes.NewValue(tftypes.Bool, nil),
		"beta":    tftypes.NewValue(tftypes.Number, nil),
		"gamma":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"delta":   tftypes.NewValue(testType.AttributeTypes["delta"], nil),
		"epsilon": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"zeta":    tftypes.NewValue(tftypes.String, nil),
	})

	testCases := map[string]struct {
		attributePath   path.Path
		expected        tftypes.Value
		expectedApplied bool
		expectedDiags   diag.Diagnostics
	}{
		"applies bool default": {
			attributePath: path.Root("alpha"),
			expected: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha":   tftypes.NewValue(tftypes.Bool, true),
				"beta":    tftypes.NewValue(tftypes.Number, nil),
				"gamma":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"delta":   tftypes.NewValue(testType.AttributeTypes["delta"], nil),
				"epsilon": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
				"zeta":    tftypes.NewValue(tftypes.String, nil),
			}),
			expectedApplied: true,
		},
		"applies number default": {
			attributePath: path.Root("beta"),
			expected: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha":   tftypes.NewValue(tftypes.Bool, nil),
				"beta":    tftypes.NewValue(tftypes.Number, big.NewFloat(42)),
				"gamma":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"delta":   tftypes.NewValue(testType.AttributeTypes["delta"], nil),
				"epsilon": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
				"zeta":    tftypes.NewValue(tftypes.String, nil),
			}),
			expectedApplied: true,
		},
		"applies list default": {
			attributePath: path.Root("gamma"),
			expected: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha": tftypes.NewValue(tftypes.Bool, nil),
				"beta":  tftypes.NewValue(tftypes.Number, nil),
				"gamma": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "alpha"),
					tftypes.NewValue(tftypes.String, "beta"),
				}),
				"delta":   tftypes.NewValue(testType.AttributeTypes["delta"], nil),
				"epsilon": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
				"zeta":    tftypes.NewValue(tftypes.String, nil),
			}),
			expectedApplied: true,
		},
		"applies object default": {
			attributePath: path.Root("delta"),
			expected: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha": tftypes.NewValue(tftypes.Bool, nil),
				"beta":  tftypes.NewValue(tftypes.Number, nil),
				"gamma": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"delta": tftypes.NewValue(testType.AttributeTypes["delta"], map[string]tftypes.Value{
					"enabled": tftypes.NewValue(tftypes.Bool, true),
				}),
				"epsilon": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
				"zeta":    tftypes.NewValue(tftypes.String, nil),
			}),
			expectedApplied: true,
		},
		"applies dynamic default": {
			attributePath: path.Root("epsilon"),
			expected: tftypes.NewValue(testType, map[string]tftypes.Value{
				"alpha":   tftypes.NewValue(tftypes.Bool, nil),
				"beta":    tftypes.NewValue(tftypes.Number, nil),
				"gamma":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"delta":   tftypes.NewValue(testType.AttributeTypes["delta"], nil),
				"epsilon": tftypes.NewValue(tftypes.String, "dynamic-default"),
				"zeta":    tftypes.NewValue(tftypes.String, nil),
			}),
			expectedApplied: true,
		},
		"returns diagnostics and skips apply when default errors": {
			attributePath:   path.Root("zeta"),
			expected:        baseConfig,
			expectedApplied: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("default failed", "string default diagnostic"),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, applied, gotDiags := setDefaultValueAtPath(t.Context(), baseConfig, testSchema, testCase.attributePath, nil)

			if applied != testCase.expectedApplied {
				t.Fatalf("unexpected applied value: got %t want %t", applied, testCase.expectedApplied)
			}

			if diff := cmp.Diff(testCase.expected, got); diff != "" {
				t.Fatalf("unexpected config diff: %s", diff)
			}

			if diff := cmp.Diff(testCase.expectedDiags, gotDiags); diff != "" {
				t.Fatalf("unexpected diagnostics diff: %s", diff)
			}
		})
	}
}
