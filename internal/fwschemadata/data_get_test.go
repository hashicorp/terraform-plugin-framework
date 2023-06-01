// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	intreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataGet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data          fwschemadata.Data
		target        any
		expected      any
		expectedDiags diag.Diagnostics
	}{
		"invalid-target": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     testtypes.StringType{},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, "test"),
					},
				),
			},
			target:   new(bool),
			expected: new(bool),
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Empty(),
					intreflect.DiagIntoIncompatibleType{
						Val: tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"string": tftypes.NewValue(tftypes.String, "test"),
							},
						),
						TargetType: reflect.TypeOf(false),
						Err: fmt.Errorf("can't unmarshal %s into *bool, expected boolean", tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"string": tftypes.String,
							},
						}),
					},
				),
			},
		},
		"invalid-type": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
			},
			target: new(struct {
				String types.String `tfsdk:"bool"`
			}),
			expected: &struct {
				String types.String `tfsdk:"bool"`
			}{
				String: types.String{},
			},
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Root("bool"),
					intreflect.DiagNewAttributeValueIntoWrongType{
						ValType:    reflect.TypeOf(types.Bool{}),
						TargetType: reflect.TypeOf(types.String{}),
						SchemaType: types.BoolType,
					},
				),
			},
		},
		"AttrTypeWithValidateError": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, "test"),
					},
				),
			},
			target: new(struct {
				String testtypes.String `tfsdk:"string"`
			}),
			expected: &struct {
				String testtypes.String `tfsdk:"string"`
			}{
				String: testtypes.String{
					CreatedBy:      testtypes.StringTypeWithValidateError{},
					InternalString: types.StringNull(),
				},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Root("string")),
			},
		},
		"AttrTypeWithValidateWarning": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, "test"),
					},
				),
			},
			target: new(struct {
				String testtypes.String `tfsdk:"string"`
			}),
			expected: &struct {
				String testtypes.String `tfsdk:"string"`
			}{
				String: testtypes.String{
					CreatedBy:      testtypes.StringTypeWithValidateWarning{},
					InternalString: types.StringValue("test"),
				},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("string")),
			},
		},
		"multiple-attributes": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"one": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
						"two": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"one": tftypes.String,
							"two": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"one": tftypes.NewValue(tftypes.String, "value1"),
						"two": tftypes.NewValue(tftypes.String, "value2"),
					},
				),
			},
			target: new(struct {
				One types.String `tfsdk:"one"`
				Two types.String `tfsdk:"two"`
			}),
			expected: &struct {
				One types.String `tfsdk:"one"`
				Two types.String `tfsdk:"two"`
			}{
				One: types.StringValue("value1"),
				Two: types.StringValue("value2"),
			},
		},
		"BoolType-types.Bool-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
			},
			target: new(struct {
				Bool types.Bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool types.Bool `tfsdk:"bool"`
			}{
				Bool: types.BoolNull(),
			},
		},
		"BoolType-types.Bool-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Bool types.Bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool types.Bool `tfsdk:"bool"`
			}{
				Bool: types.BoolUnknown(),
			},
		},
		"BoolType-types.Bool-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, true),
					},
				),
			},
			target: new(struct {
				Bool types.Bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool types.Bool `tfsdk:"bool"`
			}{
				Bool: types.BoolValue(true),
			},
		},
		"BoolType-*bool-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
			},
			target: new(struct {
				Bool *bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool *bool `tfsdk:"bool"`
			}{
				Bool: nil,
			},
		},
		"BoolType-*bool-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Bool *bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool *bool `tfsdk:"bool"`
			}{
				Bool: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("bool"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: bool\nTarget Type: *bool\nSuggested Type: basetypes.BoolValue",
				),
			},
		},
		"BoolType-*bool-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, true),
					},
				),
			},
			target: new(struct {
				Bool *bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool *bool `tfsdk:"bool"`
			}{
				Bool: pointer(true),
			},
		},
		"BoolType-bool-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
			},
			target: new(struct {
				Bool bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool bool `tfsdk:"bool"`
			}{
				Bool: false,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("bool"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received null value, however the target type cannot handle null values. Use the corresponding `types` package type, a pointer type or a custom type that handles null values.\n\n"+
						"Path: bool\nTarget Type: bool\nSuggested `types` Type: basetypes.BoolValue\nSuggested Pointer Type: *bool",
				),
			},
		},
		"BoolType-bool-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Bool bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool bool `tfsdk:"bool"`
			}{
				Bool: false,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("bool"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: bool\nTarget Type: bool\nSuggested Type: basetypes.BoolValue",
				),
			},
		},
		"BoolType-bool-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, true),
					},
				),
			},
			target: new(struct {
				Bool bool `tfsdk:"bool"`
			}),
			expected: &struct {
				Bool bool `tfsdk:"bool"`
			}{
				Bool: true,
			},
		},
		"Float64Type-types.Float64-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			target: new(struct {
				Float64 types.Float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 types.Float64 `tfsdk:"float64"`
			}{
				Float64: types.Float64Null(),
			},
		},
		"Float64Type-types.Float64-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Float64 types.Float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 types.Float64 `tfsdk:"float64"`
			}{
				Float64: types.Float64Unknown(),
			},
		},
		"Float64Type-types.Float64-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, 1.2),
					},
				),
			},
			target: new(struct {
				Float64 types.Float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 types.Float64 `tfsdk:"float64"`
			}{
				Float64: types.Float64Value(1.2),
			},
		},
		"Float64Type-*float64-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			target: new(struct {
				Float64 *float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 *float64 `tfsdk:"float64"`
			}{
				Float64: nil,
			},
		},
		"Float64Type-*float64-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Float64 *float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 *float64 `tfsdk:"float64"`
			}{
				Float64: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("float64"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: float64\nTarget Type: *float64\nSuggested Type: basetypes.Float64Value",
				),
			},
		},
		"Float64Type-*float64-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, 1.2),
					},
				),
			},
			target: new(struct {
				Float64 *float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 *float64 `tfsdk:"float64"`
			}{
				Float64: pointer(1.2),
			},
		},
		"Float64Type-float64-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			target: new(struct {
				Float64 float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 float64 `tfsdk:"float64"`
			}{
				Float64: 0.0,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("float64"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received null value, however the target type cannot handle null values. Use the corresponding `types` package type, a pointer type or a custom type that handles null values.\n\n"+
						"Path: float64\nTarget Type: float64\nSuggested `types` Type: basetypes.Float64Value\nSuggested Pointer Type: *float64",
				),
			},
		},
		"Float64Type-float64-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Float64 float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 float64 `tfsdk:"float64"`
			}{
				Float64: 0.0,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("float64"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: float64\nTarget Type: float64\nSuggested Type: basetypes.Float64Value",
				),
			},
		},
		"Float64Type-float64-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"float64": testschema.Attribute{
							Optional: true,
							Type:     types.Float64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"float64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"float64": tftypes.NewValue(tftypes.Number, 1.2),
					},
				),
			},
			target: new(struct {
				Float64 float64 `tfsdk:"float64"`
			}),
			expected: &struct {
				Float64 float64 `tfsdk:"float64"`
			}{
				Float64: 1.2,
			},
		},
		"Int64Type-types.Int64-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			target: new(struct {
				Int64 types.Int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 types.Int64 `tfsdk:"int64"`
			}{
				Int64: types.Int64Null(),
			},
		},
		"Int64Type-types.Int64-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Int64 types.Int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 types.Int64 `tfsdk:"int64"`
			}{
				Int64: types.Int64Unknown(),
			},
		},
		"Int64Type-types.Int64-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, 12),
					},
				),
			},
			target: new(struct {
				Int64 types.Int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 types.Int64 `tfsdk:"int64"`
			}{
				Int64: types.Int64Value(12),
			},
		},
		"Int64Type-*int64-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			target: new(struct {
				Int64 *int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 *int64 `tfsdk:"int64"`
			}{
				Int64: nil,
			},
		},
		"Int64Type-*int64-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Int64 *int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 *int64 `tfsdk:"int64"`
			}{
				Int64: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("int64"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: int64\nTarget Type: *int64\nSuggested Type: basetypes.Int64Value",
				),
			},
		},
		"Int64Type-*int64-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, 12),
					},
				),
			},
			target: new(struct {
				Int64 *int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 *int64 `tfsdk:"int64"`
			}{
				Int64: pointer(int64(12)),
			},
		},
		"Int64Type-int64-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, nil),
					},
				),
			},
			target: new(struct {
				Int64 int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 int64 `tfsdk:"int64"`
			}{
				Int64: 0.0,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("int64"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received null value, however the target type cannot handle null values. Use the corresponding `types` package type, a pointer type or a custom type that handles null values.\n\n"+
						"Path: int64\nTarget Type: int64\nSuggested `types` Type: basetypes.Int64Value\nSuggested Pointer Type: *int64",
				),
			},
		},
		"Int64Type-int64-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				Int64 int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 int64 `tfsdk:"int64"`
			}{
				Int64: 0,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("int64"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: int64\nTarget Type: int64\nSuggested Type: basetypes.Int64Value",
				),
			},
		},
		"Int64Type-int64-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"int64": testschema.Attribute{
							Optional: true,
							Type:     types.Int64Type,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"int64": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"int64": tftypes.NewValue(tftypes.Number, 12),
					},
				),
			},
			target: new(struct {
				Int64 int64 `tfsdk:"int64"`
			}),
			expected: &struct {
				Int64 int64 `tfsdk:"int64"`
			}{
				Int64: 12,
			},
		},
		"ListBlock-types.List-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"ListBlock-types.List-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListUnknown(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"ListBlock-types.List-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("test1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("test2"),
							},
						),
					},
				),
			},
		},
		"ListBlock-[]types.Object-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List []types.Object `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.Object `tfsdk:"list"`
			}{
				List: nil,
			},
		},
		"ListBlock-[]types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List []types.Object `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.Object `tfsdk:"list"`
			}{
				List: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: list\nTarget Type: []basetypes.ObjectValue\nSuggested Type: basetypes.ListValue",
				),
			},
		},
		"ListBlock-[]types.Object-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				List []types.Object `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.Object `tfsdk:"list"`
			}{
				List: []types.Object{
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("test1"),
						},
					),
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("test2"),
						},
					),
				},
			},
		},
		"ListBlock-[]struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}),
			expected: &struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}{
				List: nil,
			},
		},
		"ListBlock-[]struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}),
			expected: &struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}{
				List: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: list\nTarget Type: []struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.ListValue",
				),
			},
		},
		"ListBlock-[]struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}),
			expected: &struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}{
				List: []struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					{NestedString: types.StringValue("test1")},
					{NestedString: types.StringValue("test2")},
				},
			},
		},
		"ListNestedAttributes-types.List-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"ListNestedAttributes-types.List-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListUnknown(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"ListNestedAttributes-types.List-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("test1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("test2"),
							},
						),
					},
				),
			},
		},
		"ListNestedAttributes-[]types.Object-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List []types.Object `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.Object `tfsdk:"list"`
			}{
				List: nil,
			},
		},
		"ListNestedAttributes-[]types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List []types.Object `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.Object `tfsdk:"list"`
			}{
				List: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: list\nTarget Type: []basetypes.ObjectValue\nSuggested Type: basetypes.ListValue",
				),
			},
		},
		"ListNestedAttributes-[]types.Object-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				List []types.Object `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.Object `tfsdk:"list"`
			}{
				List: []types.Object{
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("test1"),
						},
					),
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("test2"),
						},
					),
				},
			},
		},
		"ListNestedAttributes-[]struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}),
			expected: &struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}{
				List: nil,
			},
		},
		"ListNestedAttributes-[]struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}),
			expected: &struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}{
				List: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: list\nTarget Type: []struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.ListValue",
				),
			},
		},
		"ListNestedAttributes-[]struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}),
			expected: &struct {
				List []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"list"`
			}{
				List: []struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					{NestedString: types.StringValue("test1")},
					{NestedString: types.StringValue("test2")},
				},
			},
		},
		"ListType-types.List-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListNull(types.StringType),
			},
		},
		"ListType-types.List-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListUnknown(types.StringType),
			},
		},
		"ListType-types.List-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							},
						),
					},
				),
			},
			target: new(struct {
				List types.List `tfsdk:"list"`
			}),
			expected: &struct {
				List types.List `tfsdk:"list"`
			}{
				List: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("test1"),
						types.StringValue("test2"),
					},
				),
			},
		},
		"ListType-[]types.String-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List []types.String `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.String `tfsdk:"list"`
			}{
				List: nil,
			},
		},
		"ListType-[]types.String-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List []types.String `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.String `tfsdk:"list"`
			}{
				List: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: list\nTarget Type: []basetypes.StringValue\nSuggested Type: basetypes.ListValue",
				),
			},
		},
		"ListType-[]types.String-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							},
						),
					},
				),
			},
			target: new(struct {
				List []types.String `tfsdk:"list"`
			}),
			expected: &struct {
				List []types.String `tfsdk:"list"`
			}{
				List: []types.String{
					types.StringValue("test1"),
					types.StringValue("test2"),
				},
			},
		},
		"ListType-[]string-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				List []string `tfsdk:"list"`
			}),
			expected: &struct {
				List []string `tfsdk:"list"`
			}{
				List: nil,
			},
		},
		"ListType-[]string-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				List []string `tfsdk:"list"`
			}),
			expected: &struct {
				List []string `tfsdk:"list"`
			}{
				List: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("list"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: list\nTarget Type: []string\nSuggested Type: basetypes.ListValue",
				),
			},
		},
		"ListType-[]string-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							},
						),
					},
				),
			},
			target: new(struct {
				List []string `tfsdk:"list"`
			}),
			expected: &struct {
				List []string `tfsdk:"list"`
			}{
				List: []string{
					"test1",
					"test2",
				},
			},
		},
		"MapNestedAttributes-types.Map-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Map types.Map `tfsdk:"map"`
			}),
			expected: &struct {
				Map types.Map `tfsdk:"map"`
			}{
				Map: types.MapNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"MapNestedAttributes-types.Map-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Map types.Map `tfsdk:"map"`
			}),
			expected: &struct {
				Map types.Map `tfsdk:"map"`
			}{
				Map: types.MapUnknown(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"MapNestedAttributes-types.Map-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"key1": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "value1"),
									},
								),
								"key2": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "value2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Map types.Map `tfsdk:"map"`
			}),
			expected: &struct {
				Map types.Map `tfsdk:"map"`
			}{
				Map: types.MapValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
					map[string]attr.Value{
						"key1": types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("value1"),
							},
						),
						"key2": types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("value2"),
							},
						),
					},
				),
			},
		},
		"MapNestedAttributes-map[string]types.Object-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Map map[string]types.Object `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]types.Object `tfsdk:"map"`
			}{
				Map: nil,
			},
		},
		"MapNestedAttributes-map[string]types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Map map[string]types.Object `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]types.Object `tfsdk:"map"`
			}{
				Map: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("map"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: map\nTarget Type: map[string]basetypes.ObjectValue\nSuggested Type: basetypes.MapValue",
				),
			},
		},
		"MapNestedAttributes-map[string]types.Object-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"key1": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "value1"),
									},
								),
								"key2": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "value2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Map map[string]types.Object `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]types.Object `tfsdk:"map"`
			}{
				Map: map[string]types.Object{
					"key1": types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("value1"),
						},
					),
					"key2": types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("value2"),
						},
					),
				},
			},
		},
		"MapNestedAttributes-map[string]struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Map map[string]struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"map"`
			}{
				Map: nil,
			},
		},
		"MapNestedAttributes-map[string]struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Map map[string]struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"map"`
			}{
				Map: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("map"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: map\nTarget Type: map[string]struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.MapValue",
				),
			},
		},
		"MapNestedAttributes-map[string]struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							map[string]tftypes.Value{
								"key1": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "value1"),
									},
								),
								"key2": tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "value2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Map map[string]struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"map"`
			}{
				Map: map[string]struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					"key1": {NestedString: types.StringValue("value1")},
					"key2": {NestedString: types.StringValue("value2")},
				},
			},
		},
		"MapType-types.Map-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Map types.Map `tfsdk:"map"`
			}),
			expected: &struct {
				Map types.Map `tfsdk:"map"`
			}{
				Map: types.MapNull(types.StringType),
			},
		},
		"MapType-types.Map-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Map types.Map `tfsdk:"map"`
			}),
			expected: &struct {
				Map types.Map `tfsdk:"map"`
			}{
				Map: types.MapUnknown(types.StringType),
			},
		},
		"MapType-types.Map-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							map[string]tftypes.Value{
								"key1": tftypes.NewValue(tftypes.String, "value1"),
								"key2": tftypes.NewValue(tftypes.String, "value2"),
							},
						),
					},
				),
			},
			target: new(struct {
				Map types.Map `tfsdk:"map"`
			}),
			expected: &struct {
				Map types.Map `tfsdk:"map"`
			}{
				Map: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"key1": types.StringValue("value1"),
						"key2": types.StringValue("value2"),
					},
				),
			},
		},
		"MapType-map[string]types.String-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Map map[string]types.String `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]types.String `tfsdk:"map"`
			}{
				Map: nil,
			},
		},
		"MapType-map[string]types.String-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Map map[string]types.String `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]types.String `tfsdk:"map"`
			}{
				Map: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("map"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: map\nTarget Type: map[string]basetypes.StringValue\nSuggested Type: basetypes.MapValue",
				),
			},
		},
		"MapType-map[string]types.String-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							map[string]tftypes.Value{
								"key1": tftypes.NewValue(tftypes.String, "value1"),
								"key2": tftypes.NewValue(tftypes.String, "value2"),
							},
						),
					},
				),
			},
			target: new(struct {
				Map map[string]types.String `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]types.String `tfsdk:"map"`
			}{
				Map: map[string]types.String{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			},
		},
		"MapType-map[string]string-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Map map[string]string `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]string `tfsdk:"map"`
			}{
				Map: nil,
			},
		},
		"MapType-map[string]string-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Map map[string]string `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]string `tfsdk:"map"`
			}{
				Map: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("map"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: map\nTarget Type: map[string]string\nSuggested Type: basetypes.MapValue",
				),
			},
		},
		"MapType-map[string]string-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"map": testschema.Attribute{
							Optional: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map": tftypes.Map{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"map": tftypes.NewValue(
							tftypes.Map{
								ElementType: tftypes.String,
							},
							map[string]tftypes.Value{
								"key1": tftypes.NewValue(tftypes.String, "value1"),
								"key2": tftypes.NewValue(tftypes.String, "value2"),
							},
						),
					},
				),
			},
			target: new(struct {
				Map map[string]string `tfsdk:"map"`
			}),
			expected: &struct {
				Map map[string]string `tfsdk:"map"`
			}{
				Map: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"ObjectType-types.Object-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectNull(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
				),
			},
		},
		"ObjectType-types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectUnknown(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
				),
			},
		},
		"ObjectType-types.Object-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
					map[string]attr.Value{
						"nested_string": types.StringValue("test1"),
					},
				),
			},
		},
		"ObjectType-*struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: nil,
			},
		},
		"ObjectType-*struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: object\nTarget Type: *struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.ObjectValue",
				),
			},
		},
		"ObjectType-*struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: &struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.StringValue("test1"),
				},
			},
		},
		"ObjectType-struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.String{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received null value, however the target type cannot handle null values. Use the corresponding `types` package type, a pointer type or a custom type that handles null values.\n\n"+
						"Path: object\nTarget Type: struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested `types` Type: basetypes.ObjectValue\nSuggested Pointer Type: *struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }",
				),
			},
		},
		"ObjectType-struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.String{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: object\nTarget Type: struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.ObjectValue",
				),
			},
		},
		"ObjectType-struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.Attribute{
							Optional: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_string": types.StringType,
								},
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.StringValue("test1"),
				},
			},
		},
		"SetBlock-types.Set-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"SetBlock-types.Set-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetUnknown(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"SetBlock-types.Set-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("test1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("test2"),
							},
						),
					},
				),
			},
		},
		"SetBlock-[]types.Object-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set []types.Object `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.Object `tfsdk:"set"`
			}{
				Set: nil,
			},
		},
		"SetBlock-[]types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set []types.Object `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.Object `tfsdk:"set"`
			}{
				Set: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: set\nTarget Type: []basetypes.ObjectValue\nSuggested Type: basetypes.SetValue",
				),
			},
		},
		"SetBlock-[]types.Object-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Set []types.Object `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.Object `tfsdk:"set"`
			}{
				Set: []types.Object{
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("test1"),
						},
					),
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("test2"),
						},
					),
				},
			},
		},
		"SetBlock-[]struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}),
			expected: &struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}{
				Set: nil,
			},
		},
		"SetBlock-[]struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}),
			expected: &struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}{
				Set: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: set\nTarget Type: []struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.SetValue",
				),
			},
		},
		"SetBlock-[]struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}),
			expected: &struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}{
				Set: []struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					{NestedString: types.StringValue("test1")},
					{NestedString: types.StringValue("test2")},
				},
			},
		},
		"SetNestedAttributes-types.Set-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"SetNestedAttributes-types.Set-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetUnknown(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
				),
			},
		},
		"SetNestedAttributes-types.Set-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
					},
					[]attr.Value{
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("test1"),
							},
						),
						types.ObjectValueMust(
							map[string]attr.Type{
								"nested_string": types.StringType,
							},
							map[string]attr.Value{
								"nested_string": types.StringValue("test2"),
							},
						),
					},
				),
			},
		},
		"SetNestedAttributes-[]types.Object-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set []types.Object `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.Object `tfsdk:"set"`
			}{
				Set: nil,
			},
		},
		"SetNestedAttributes-[]types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set []types.Object `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.Object `tfsdk:"set"`
			}{
				Set: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: set\nTarget Type: []basetypes.ObjectValue\nSuggested Type: basetypes.SetValue",
				),
			},
		},
		"SetNestedAttributes-[]types.Object-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Set []types.Object `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.Object `tfsdk:"set"`
			}{
				Set: []types.Object{
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("test1"),
						},
					),
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_string": types.StringType,
						},
						map[string]attr.Value{
							"nested_string": types.StringValue("test2"),
						},
					),
				},
			},
		},
		"SetNestedAttributes-[]struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}),
			expected: &struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}{
				Set: nil,
			},
		},
		"SetNestedAttributes-[]struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}),
			expected: &struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}{
				Set: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: set\nTarget Type: []struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.SetValue",
				),
			},
		},
		"SetNestedAttributes-[]struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_string": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test1"),
									},
								),
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_string": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"nested_string": tftypes.NewValue(tftypes.String, "test2"),
									},
								),
							},
						),
					},
				),
			},
			target: new(struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}),
			expected: &struct {
				Set []struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"set"`
			}{
				Set: []struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					{NestedString: types.StringValue("test1")},
					{NestedString: types.StringValue("test2")},
				},
			},
		},
		"SetType-types.Set-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetNull(types.StringType),
			},
		},
		"SetType-types.Set-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetUnknown(types.StringType),
			},
		},
		"SetType-types.Set-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							},
						),
					},
				),
			},
			target: new(struct {
				Set types.Set `tfsdk:"set"`
			}),
			expected: &struct {
				Set types.Set `tfsdk:"set"`
			}{
				Set: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("test1"),
						types.StringValue("test2"),
					},
				),
			},
		},
		"SetType-[]types.String-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set []types.String `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.String `tfsdk:"set"`
			}{
				Set: nil,
			},
		},
		"SetType-[]types.String-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set []types.String `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.String `tfsdk:"set"`
			}{
				Set: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: set\nTarget Type: []basetypes.StringValue\nSuggested Type: basetypes.SetValue",
				),
			},
		},
		"SetType-[]types.String-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							},
						),
					},
				),
			},
			target: new(struct {
				Set []types.String `tfsdk:"set"`
			}),
			expected: &struct {
				Set []types.String `tfsdk:"set"`
			}{
				Set: []types.String{
					types.StringValue("test1"),
					types.StringValue("test2"),
				},
			},
		},
		"SetType-[]string-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Set []string `tfsdk:"set"`
			}),
			expected: &struct {
				Set []string `tfsdk:"set"`
			}{
				Set: nil,
			},
		},
		"SetType-[]string-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Set []string `tfsdk:"set"`
			}),
			expected: &struct {
				Set []string `tfsdk:"set"`
			}{
				Set: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("set"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: set\nTarget Type: []string\nSuggested Type: basetypes.SetValue",
				),
			},
		},
		"SetType-[]string-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "test1"),
								tftypes.NewValue(tftypes.String, "test2"),
							},
						),
					},
				),
			},
			target: new(struct {
				Set []string `tfsdk:"set"`
			}),
			expected: &struct {
				Set []string `tfsdk:"set"`
			}{
				Set: []string{
					"test1",
					"test2",
				},
			},
		},
		"SingleBlock-types.Object-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectNull(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
				),
			},
		},
		"SingleBlock-types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectUnknown(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
				),
			},
		},
		"SingleBlock-types.Object-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
					map[string]attr.Value{
						"nested_string": types.StringValue("test1"),
					},
				),
			},
		},
		"SingleBlock-*struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: nil,
			},
		},
		"SingleBlock-*struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: object\nTarget Type: *struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.ObjectValue",
				),
			},
		},
		"SingleBlock-*struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: &struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.StringValue("test1"),
				},
			},
		},
		"SingleBlock-struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.String{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received null value, however the target type cannot handle null values. Use the corresponding `types` package type, a pointer type or a custom type that handles null values.\n\n"+
						"Path: object\nTarget Type: struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested `types` Type: basetypes.ObjectValue\nSuggested Pointer Type: *struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }",
				),
			},
		},
		"SingleBlock-struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.String{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: object\nTarget Type: struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.ObjectValue",
				),
			},
		},
		"SingleBlock-struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"object": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.StringValue("test1"),
				},
			},
		},
		"SingleNestedAttributes-types.Object-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectNull(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
				),
			},
		},
		"SingleNestedAttributes-types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectUnknown(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
				),
			},
		},
		"SingleNestedAttributes-types.Object-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object types.Object `tfsdk:"object"`
			}),
			expected: &struct {
				Object types.Object `tfsdk:"object"`
			}{
				Object: types.ObjectValueMust(
					map[string]attr.Type{
						"nested_string": types.StringType,
					},
					map[string]attr.Value{
						"nested_string": types.StringValue("test1"),
					},
				),
			},
		},
		"SingleNestedAttributes-*struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: nil,
			},
		},
		"SingleNestedAttributes-*struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: object\nTarget Type: *struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.ObjectValue",
				),
			},
		},
		"SingleNestedAttributes-*struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object *struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: &struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.StringValue("test1"),
				},
			},
		},
		"SingleNestedAttributes-struct-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.String{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received null value, however the target type cannot handle null values. Use the corresponding `types` package type, a pointer type or a custom type that handles null values.\n\n"+
						"Path: object\nTarget Type: struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested `types` Type: basetypes.ObjectValue\nSuggested Pointer Type: *struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }",
				),
			},
		},
		"SingleNestedAttributes-struct-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.String{},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("object"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: object\nTarget Type: struct { NestedString basetypes.StringValue \"tfsdk:\\\"nested_string\\\"\" }\nSuggested Type: basetypes.ObjectValue",
				),
			},
		},
		"SingleNestedAttributes-struct-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"object": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"nested_string": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"object": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"object": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"nested_string": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"nested_string": tftypes.NewValue(tftypes.String, "test1"),
							},
						),
					},
				),
			},
			target: new(struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}),
			expected: &struct {
				Object struct {
					NestedString types.String `tfsdk:"nested_string"`
				} `tfsdk:"object"`
			}{
				Object: struct {
					NestedString types.String `tfsdk:"nested_string"`
				}{
					NestedString: types.StringValue("test1"),
				},
			},
		},
		"StringType-types.string-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
			target: new(struct {
				String types.String `tfsdk:"string"`
			}),
			expected: &struct {
				String types.String `tfsdk:"string"`
			}{
				String: types.StringNull(),
			},
		},
		"StringType-types.string-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				String types.String `tfsdk:"string"`
			}),
			expected: &struct {
				String types.String `tfsdk:"string"`
			}{
				String: types.StringUnknown(),
			},
		},
		"StringType-types.string-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, "test"),
					},
				),
			},
			target: new(struct {
				String types.String `tfsdk:"string"`
			}),
			expected: &struct {
				String types.String `tfsdk:"string"`
			}{
				String: types.StringValue("test"),
			},
		},
		"StringType-*string-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
			target: new(struct {
				String *string `tfsdk:"string"`
			}),
			expected: &struct {
				String *string `tfsdk:"string"`
			}{
				String: nil,
			},
		},
		"StringType-*string-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				String *string `tfsdk:"string"`
			}),
			expected: &struct {
				String *string `tfsdk:"string"`
			}{
				String: nil,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("string"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: string\nTarget Type: *string\nSuggested Type: basetypes.StringValue",
				),
			},
		},
		"StringType-*string-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, "test"),
					},
				),
			},
			target: new(struct {
				String *string `tfsdk:"string"`
			}),
			expected: &struct {
				String *string `tfsdk:"string"`
			}{
				String: pointer("test"),
			},
		},
		"StringType-string-null": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, nil),
					},
				),
			},
			target: new(struct {
				String string `tfsdk:"string"`
			}),
			expected: &struct {
				String string `tfsdk:"string"`
			}{
				String: "",
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("string"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received null value, however the target type cannot handle null values. Use the corresponding `types` package type, a pointer type or a custom type that handles null values.\n\n"+
						"Path: string\nTarget Type: string\nSuggested `types` Type: basetypes.StringValue\nSuggested Pointer Type: *string",
				),
			},
		},
		"StringType-string-unknown": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					},
				),
			},
			target: new(struct {
				String string `tfsdk:"string"`
			}),
			expected: &struct {
				String string `tfsdk:"string"`
			}{
				String: "",
			},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("string"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to build a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Received unknown value, however the target type cannot handle unknown values. Use the corresponding `types` package type or a custom type that handles unknown values.\n\n"+
						"Path: string\nTarget Type: string\nSuggested Type: basetypes.StringValue",
				),
			},
		},
		"StringType-string-value": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, "test"),
					},
				),
			},
			target: new(struct {
				String string `tfsdk:"string"`
			}),
			expected: &struct {
				String string `tfsdk:"string"`
			}{
				String: "test",
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.data.Get(context.Background(), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
