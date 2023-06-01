// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"fmt"
	"math/big"
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

func TestDataGetAtPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data          fwschemadata.Data
		path          path.Path
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
			path:     path.Root("string"),
			target:   new(bool),
			expected: new(bool),
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Root("string"),
					intreflect.DiagIntoIncompatibleType{
						Val:        tftypes.NewValue(tftypes.String, "test"),
						TargetType: reflect.TypeOf(false),
						Err:        fmt.Errorf("can't unmarshal %s into *bool, expected boolean", tftypes.String),
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
			path:     path.Root("bool"),
			target:   new(types.String),
			expected: &types.String{},
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
			path:   path.Root("string"),
			target: new(testtypes.String),
			expected: &testtypes.String{
				CreatedBy:      testtypes.StringTypeWithValidateError{},
				InternalString: types.StringNull(),
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
			path:   path.Root("string"),
			target: new(testtypes.String),
			expected: &testtypes.String{
				CreatedBy:      testtypes.StringTypeWithValidateWarning{},
				InternalString: types.StringValue("test"),
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("string")),
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
			path:     path.Root("bool"),
			target:   new(types.Bool),
			expected: pointer(types.BoolNull()),
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
			path:     path.Root("bool"),
			target:   new(types.Bool),
			expected: pointer(types.BoolUnknown()),
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
			path:     path.Root("bool"),
			target:   new(types.Bool),
			expected: pointer(types.BoolValue(true)),
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
			path:     path.Root("bool"),
			target:   new(*bool),
			expected: new(*bool),
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
			path:     path.Root("bool"),
			target:   new(*bool),
			expected: new(*bool),
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
			path:     path.Root("bool"),
			target:   new(*bool),
			expected: pointer(pointer(true)),
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
			path:     path.Root("bool"),
			target:   new(bool),
			expected: new(bool),
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
			path:     path.Root("bool"),
			target:   new(bool),
			expected: pointer(false),
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
			path:     path.Root("bool"),
			target:   new(bool),
			expected: pointer(true),
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
			path:     path.Root("float64"),
			target:   new(types.Float64),
			expected: pointer(types.Float64Null()),
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
			path:     path.Root("float64"),
			target:   new(types.Float64),
			expected: pointer(types.Float64Unknown()),
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
			path:     path.Root("float64"),
			target:   new(types.Float64),
			expected: pointer(types.Float64Value(1.2)),
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
			path:     path.Root("float64"),
			target:   new(*float64),
			expected: new(*float64),
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
			path:     path.Root("float64"),
			target:   new(*float64),
			expected: new(*float64),
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
			path:     path.Root("float64"),
			target:   new(*float64),
			expected: pointer(pointer(1.2)),
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
			path:     path.Root("float64"),
			target:   new(float64),
			expected: new(float64),
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
			path:     path.Root("float64"),
			target:   new(float64),
			expected: new(float64),
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
			path:     path.Root("float64"),
			target:   new(float64),
			expected: pointer(1.2),
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
			path:     path.Root("int64"),
			target:   new(types.Int64),
			expected: pointer(types.Int64Null()),
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
			path:     path.Root("int64"),
			target:   new(types.Int64),
			expected: pointer(types.Int64Unknown()),
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
			path:     path.Root("int64"),
			target:   new(types.Int64),
			expected: pointer(types.Int64Value(12)),
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
			path:     path.Root("int64"),
			target:   new(*int64),
			expected: new(*int64),
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
			path:     path.Root("int64"),
			target:   new(*int64),
			expected: new(*int64),
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
			path:     path.Root("int64"),
			target:   new(*int64),
			expected: pointer(pointer(int64(12))),
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
			path:     path.Root("int64"),
			target:   new(int64),
			expected: new(int64),
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
			path:     path.Root("int64"),
			target:   new(int64),
			expected: new(int64),
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
			path:     path.Root("int64"),
			target:   new(int64),
			expected: pointer(int64(12)),
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: pointer(types.ListNull(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: pointer(types.ListUnknown(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: pointer(types.ListValueMust(
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
			)),
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
			path:     path.Root("list"),
			target:   new([]types.Object),
			expected: new([]types.Object),
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
			path:     path.Root("list"),
			target:   new([]types.Object),
			expected: new([]types.Object),
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
			path:   path.Root("list"),
			target: new([]types.Object),
			expected: &[]types.Object{
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
			path: path.Root("list"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("list"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("list"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &[]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				{NestedString: types.StringValue("test1")},
				{NestedString: types.StringValue("test2")},
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: pointer(types.ListNull(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: pointer(types.ListUnknown(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: pointer(types.ListValueMust(
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
			)),
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
			path:     path.Root("list"),
			target:   new([]types.Object),
			expected: new([]types.Object),
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
			path:     path.Root("list"),
			target:   new([]types.Object),
			expected: new([]types.Object),
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
			path:   path.Root("list"),
			target: new([]types.Object),
			expected: &[]types.Object{
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
			path: path.Root("list"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("list"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("list"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &[]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				{NestedString: types.StringValue("test1")},
				{NestedString: types.StringValue("test2")},
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
			path:     path.Root("list"),
			target:   new(types.List),
			expected: pointer(types.ListNull(types.StringType)),
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
			path:     path.Root("list"),
			target:   new(types.List),
			expected: pointer(types.ListUnknown(types.StringType)),
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: pointer(types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("test1"),
					types.StringValue("test2"),
				},
			)),
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
			path:     path.Root("list"),
			target:   new([]types.String),
			expected: new([]types.String),
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
			path:     path.Root("list"),
			target:   new([]types.String),
			expected: new([]types.String),
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
			path:   path.Root("list"),
			target: new([]types.String),
			expected: &[]types.String{
				types.StringValue("test1"),
				types.StringValue("test2"),
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
			path:     path.Root("list"),
			target:   new([]string),
			expected: new([]string),
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
			path:     path.Root("list"),
			target:   new([]string),
			expected: new([]string),
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
			path:   path.Root("list"),
			target: new([]string),
			expected: &[]string{
				"test1",
				"test2",
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
			path:   path.Root("map"),
			target: new(types.Map),
			expected: pointer(types.MapNull(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("map"),
			target: new(types.Map),
			expected: pointer(types.MapUnknown(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("map"),
			target: new(types.Map),
			expected: pointer(types.MapValueMust(
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
			)),
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
			path:     path.Root("map"),
			target:   new(map[string]types.Object),
			expected: new(map[string]types.Object),
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
			path:     path.Root("map"),
			target:   new(map[string]types.Object),
			expected: new(map[string]types.Object),
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
			path:   path.Root("map"),
			target: new(map[string]types.Object),
			expected: &map[string]types.Object{
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
			path: path.Root("map"),
			target: new(map[string]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new(map[string]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("map"),
			target: new(map[string]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new(map[string]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("map"),
			target: new(map[string]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &map[string]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				"key1": {NestedString: types.StringValue("value1")},
				"key2": {NestedString: types.StringValue("value2")},
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
			path:     path.Root("map"),
			target:   new(types.Map),
			expected: pointer(types.MapNull(types.StringType)),
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
			path:     path.Root("map"),
			target:   new(types.Map),
			expected: pointer(types.MapUnknown(types.StringType)),
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
			path:   path.Root("map"),
			target: new(types.Map),
			expected: pointer(types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			)),
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
			path:     path.Root("map"),
			target:   new(map[string]types.String),
			expected: new(map[string]types.String),
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
			path:     path.Root("map"),
			target:   new(map[string]types.String),
			expected: new(map[string]types.String),
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
			path:   path.Root("map"),
			target: new(map[string]types.String),
			expected: &map[string]types.String{
				"key1": types.StringValue("value1"),
				"key2": types.StringValue("value2"),
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
			path:     path.Root("map"),
			target:   new(map[string]string),
			expected: new(map[string]string),
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
			path:     path.Root("map"),
			target:   new(map[string]string),
			expected: new(map[string]string),
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
			path:   path.Root("map"),
			target: new(map[string]string),
			expected: &map[string]string{
				"key1": "value1",
				"key2": "value2",
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectNull(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
			)),
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectUnknown(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
			)),
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectValueMust(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
				map[string]attr.Value{
					"nested_string": types.StringValue("test1"),
				},
			)),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: pointer(&struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.StringValue("test1"),
			}),
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.String{},
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.String{},
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.StringValue("test1"),
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: pointer(types.SetNull(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: pointer(types.SetUnknown(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: pointer(types.SetValueMust(
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
			)),
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
			path:     path.Root("set"),
			target:   new([]types.Object),
			expected: new([]types.Object),
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
			path:     path.Root("set"),
			target:   new([]types.Object),
			expected: new([]types.Object),
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
			path:   path.Root("set"),
			target: new([]types.Object),
			expected: &[]types.Object{
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
			path: path.Root("set"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("set"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("set"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &[]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				{NestedString: types.StringValue("test1")},
				{NestedString: types.StringValue("test2")},
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: pointer(types.SetNull(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: pointer(types.SetUnknown(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
			)),
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: pointer(types.SetValueMust(
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
			)),
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
			path:     path.Root("set"),
			target:   new([]types.Object),
			expected: new([]types.Object),
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
			path:     path.Root("set"),
			target:   new([]types.Object),
			expected: new([]types.Object),
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
			path:   path.Root("set"),
			target: new([]types.Object),
			expected: &[]types.Object{
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
			path: path.Root("set"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("set"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("set"),
			target: new([]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &[]struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				{NestedString: types.StringValue("test1")},
				{NestedString: types.StringValue("test2")},
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
			path:     path.Root("set"),
			target:   new(types.Set),
			expected: pointer(types.SetNull(types.StringType)),
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
			path:     path.Root("set"),
			target:   new(types.Set),
			expected: pointer(types.SetUnknown(types.StringType)),
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: pointer(types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("test1"),
					types.StringValue("test2"),
				},
			)),
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
			path:     path.Root("set"),
			target:   new([]types.String),
			expected: new([]types.String),
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
			path:     path.Root("set"),
			target:   new([]types.String),
			expected: new([]types.String),
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
			path:   path.Root("set"),
			target: new([]types.String),
			expected: &[]types.String{
				types.StringValue("test1"),
				types.StringValue("test2"),
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
			path:     path.Root("set"),
			target:   new([]string),
			expected: new([]string),
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
			path:     path.Root("set"),
			target:   new([]string),
			expected: new([]string),
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
			path:   path.Root("set"),
			target: new([]string),
			expected: &[]string{
				"test1",
				"test2",
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectNull(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
			)),
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectUnknown(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
			)),
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectValueMust(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
				map[string]attr.Value{
					"nested_string": types.StringValue("test1"),
				},
			)),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: pointer(&struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.StringValue("test1"),
			}),
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.String{},
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.String{},
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.StringValue("test1"),
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectNull(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
			)),
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectUnknown(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
			)),
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
			path:   path.Root("object"),
			target: new(types.Object),
			expected: pointer(types.ObjectValueMust(
				map[string]attr.Type{
					"nested_string": types.StringType,
				},
				map[string]attr.Value{
					"nested_string": types.StringValue("test1"),
				},
			)),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
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
			path: path.Root("object"),
			target: new(*struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: pointer(&struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.StringValue("test1"),
			}),
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.String{},
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.String{},
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
			path: path.Root("object"),
			target: new(struct {
				NestedString types.String `tfsdk:"nested_string"`
			}),
			expected: &struct {
				NestedString types.String `tfsdk:"nested_string"`
			}{
				NestedString: types.StringValue("test1"),
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
			path:     path.Root("string"),
			target:   new(types.String),
			expected: pointer(types.StringNull()),
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
			path:     path.Root("string"),
			target:   new(types.String),
			expected: pointer(types.StringUnknown()),
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
			path:     path.Root("string"),
			target:   new(types.String),
			expected: pointer(types.StringValue("test")),
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
			path:     path.Root("string"),
			target:   new(*string),
			expected: new(*string),
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
			path:     path.Root("string"),
			target:   new(*string),
			expected: new(*string),
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
			path:     path.Root("string"),
			target:   new(*string),
			expected: pointer(pointer("test")),
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
			path:     path.Root("string"),
			target:   new(string),
			expected: new(string),
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
			path:     path.Root("string"),
			target:   new(string),
			expected: new(string),
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
			path:     path.Root("string"),
			target:   new(string),
			expected: pointer("test"),
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.data.GetAtPath(context.Background(), tc.path, tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				for _, d := range diags {
					t.Log(d)
				}
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			// Prevent pointer inequality
			comparers := []cmp.Option{
				cmp.Comparer(func(i, j *big.Float) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && i.Cmp(j) == 0)
				}),
				cmp.Comparer(func(i, j *testtypes.String) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
				cmp.Comparer(func(i, j *types.Bool) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
				cmp.Comparer(func(i, j *types.Float64) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
				cmp.Comparer(func(i, j *types.Int64) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
				cmp.Comparer(func(i, j *types.List) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
				cmp.Comparer(func(i, j *types.Map) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
				cmp.Comparer(func(i, j *types.Object) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
				cmp.Comparer(func(i, j *types.Set) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
				cmp.Comparer(func(i, j *types.String) bool {
					return (i == nil && j == nil) || (i != nil && j != nil && cmp.Equal(*i, *j))
				}),
			}

			if diff := cmp.Diff(tc.target, tc.expected, comparers...); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
