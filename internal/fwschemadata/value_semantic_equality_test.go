// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValueSemanticEquality(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		"BoolValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.BoolValue(false),
				ProposedNewValue: types.BoolValue(true),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.BoolValue(true),
			},
		},
		"BoolValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(false),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(false),
					SemanticEquals: true,
				},
			},
		},
		"BoolValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(false),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: false,
				},
			},
		},
		"BoolValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(false),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"Float64Value": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.Float64Value(1.2),
				ProposedNewValue: types.Float64Value(2.4),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.Float64Value(2.4),
			},
		},
		"Float64ValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(1.2),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(2.4),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(1.2),
					SemanticEquals: true,
				},
			},
		},
		"Float64ValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(1.2),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(2.4),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(2.4),
					SemanticEquals: false,
				},
			},
		},
		"Float64ValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(1.2),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(2.4),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Float64ValueWithSemanticEquals{
					Float64Value:   types.Float64Value(2.4),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"Int64Value": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.Int64Value(12),
				ProposedNewValue: types.Int64Value(24),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.Int64Value(24),
			},
		},
		"Int64ValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(12),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(24),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(12),
					SemanticEquals: true,
				},
			},
		},
		"Int64ValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(12),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(24),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(24),
					SemanticEquals: false,
				},
			},
		},
		"Int64ValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(12),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(24),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.Int64ValueWithSemanticEquals{
					Int64Value:     types.Int64Value(24),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"ListValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
		},
		"ListValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"ListValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"ListValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"MapValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("new"),
					},
				),
			},
		},
		"MapValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"MapValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"MapValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"NumberValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.NumberValue(big.NewFloat(1.2)),
				ProposedNewValue: types.NumberValue(big.NewFloat(2.4)),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.NumberValue(big.NewFloat(2.4)),
			},
		},
		"NumberValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(1.2)),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(1.2)),
					SemanticEquals: true,
				},
			},
		},
		"NumberValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(1.2)),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: false,
				},
			},
		},
		"NumberValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(1.2)),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.NumberValueWithSemanticEquals{
					NumberValue:    types.NumberValue(big.NewFloat(2.4)),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"ObjectValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.StringType,
					},
					map[string]attr.Value{
						"test_attr": types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.StringType,
					},
					map[string]attr.Value{
						"test_attr": types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.StringType,
					},
					map[string]attr.Value{
						"test_attr": types.StringValue("new"),
					},
				),
			},
		},
		"ObjectValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"ObjectValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"ObjectValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"SetValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
		},
		"SetValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"SetValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"SetValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"StringValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.StringValue("prior"),
				ProposedNewValue: types.StringValue("new"),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.StringValue("new"),
			},
		},
		"StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("prior"),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("new"),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("prior"),
					SemanticEquals: true,
				},
			},
		},
		"StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("prior"),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("new"),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("new"),
					SemanticEquals: false,
				},
			},
		},
		"StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("prior"),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("new"),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.StringValueWithSemanticEquals{
					StringValue:    types.StringValue("new"),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testCase.request.ProposedNewValue,
			}

			fwschemadata.ValueSemanticEquality(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
