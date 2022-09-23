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
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	intreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
				InternalString: types.String{Value: ""},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Root("string")),
			},
		},
		"AttrTypeWithValidateWarning": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
				InternalString: types.String{Value: "test"},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("string")),
			},
		},
		"BoolType-types.Bool-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
			expected: &types.Bool{Null: true},
		},
		"BoolType-types.Bool-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
			expected: &types.Bool{Unknown: true},
		},
		"BoolType-types.Bool-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
			expected: &types.Bool{Value: true},
		},
		"BoolType-*bool-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
						"Received unknown value for bool, however the current struct field type *bool cannot handle unknown values. Use types.Bool, or a custom type that supports unknown values instead.",
				),
			},
		},
		"BoolType-*bool-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
						"Received null value for bool, however the current struct field type bool cannot handle null values. Use a pointer type (*bool), types.Bool, or a custom type that supports null values instead.",
				),
			},
		},
		"BoolType-bool-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
						"Received unknown value for bool, however the current struct field type bool cannot handle unknown values. Use types.Bool, or a custom type that supports unknown values instead.",
				),
			},
		},
		"BoolType-bool-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"bool": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
			expected: &types.Float64{Null: true},
		},
		"Float64Type-types.Float64-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
			expected: &types.Float64{Unknown: true},
		},
		"Float64Type-types.Float64-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
			expected: &types.Float64{Value: 1.2},
		},
		"Float64Type-*float64-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
						"Received unknown value for float64, however the current struct field type *float64 cannot handle unknown values. Use types.Float64, or a custom type that supports unknown values instead.",
				),
			},
		},
		"Float64Type-*float64-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
						"Received null value for float64, however the current struct field type float64 cannot handle null values. Use a pointer type (*float64), types.Float64, or a custom type that supports null values instead.",
				),
			},
		},
		"Float64Type-float64-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
						"Received unknown value for float64, however the current struct field type float64 cannot handle unknown values. Use types.Float64, or a custom type that supports unknown values instead.",
				),
			},
		},
		"Float64Type-float64-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"float64": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
			expected: &types.Int64{Null: true},
		},
		"Int64Type-types.Int64-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
			expected: &types.Int64{Unknown: true},
		},
		"Int64Type-types.Int64-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
			expected: &types.Int64{Value: 12},
		},
		"Int64Type-*int64-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
						"Received unknown value for int64, however the current struct field type *int64 cannot handle unknown values. Use types.Int64, or a custom type that supports unknown values instead.",
				),
			},
		},
		"Int64Type-*int64-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
						"Received null value for int64, however the current struct field type int64 cannot handle null values. Use a pointer type (*int64), types.Int64, or a custom type that supports null values instead.",
				),
			},
		},
		"Int64Type-int64-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
						"Received unknown value for int64, however the current struct field type int64 cannot handle unknown values. Use types.Int64, or a custom type that supports unknown values instead.",
				),
			},
		},
		"Int64Type-int64-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"int64": {
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
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
			expected: &types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Null: true,
			},
		},
		"ListBlock-types.List-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
			expected: &types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Unknown: true,
			},
		},
		"ListBlock-types.List-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
			expected: &types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "test1"},
						},
					},
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "test2"},
						},
					},
				},
			},
		},
		"ListBlock-[]types.Object-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
						"Received unknown value for list, however the current struct field type []types.Object cannot handle unknown values. Use types.List, or a custom type that supports unknown values instead.",
				),
			},
		},
		"ListBlock-[]types.Object-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
				{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "test1"},
					},
				},
				{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "test2"},
					},
				},
			},
		},
		"ListBlock-[]struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
						`Received unknown value for list, however the current struct field type []struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.List, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"ListBlock-[]struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"list": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
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
				{NestedString: types.String{Value: "test1"}},
				{NestedString: types.String{Value: "test2"}},
			},
		},
		"ListNestedAttributes-types.List-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Null: true,
			},
		},
		"ListNestedAttributes-types.List-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Unknown: true,
			},
		},
		"ListNestedAttributes-types.List-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "test1"},
						},
					},
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "test2"},
						},
					},
				},
			},
		},
		"ListNestedAttributes-[]types.Object-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						"Received unknown value for list, however the current struct field type []types.Object cannot handle unknown values. Use types.List, or a custom type that supports unknown values instead.",
				),
			},
		},
		"ListNestedAttributes-[]types.Object-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "test1"},
					},
				},
				{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "test2"},
					},
				},
			},
		},
		"ListNestedAttributes-[]struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						`Received unknown value for list, however the current struct field type []struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.List, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"ListNestedAttributes-[]struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				{NestedString: types.String{Value: "test1"}},
				{NestedString: types.String{Value: "test2"}},
			},
		},
		"ListType-types.List-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: &types.List{
				ElemType: types.StringType,
				Null:     true,
			},
		},
		"ListType-types.List-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
			path:   path.Root("list"),
			target: new(types.List),
			expected: &types.List{
				ElemType: types.StringType,
				Unknown:  true,
			},
		},
		"ListType-types.List-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
			expected: &types.List{
				ElemType: types.StringType,
				Elems: []attr.Value{
					types.String{Value: "test1"},
					types.String{Value: "test2"},
				},
			},
		},
		"ListType-[]types.String-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
						"Received unknown value for list, however the current struct field type []types.String cannot handle unknown values. Use types.List, or a custom type that supports unknown values instead.",
				),
			},
		},
		"ListType-[]types.String-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
				{Value: "test1"},
				{Value: "test2"},
			},
		},
		"ListType-[]string-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
						"Received unknown value for list, however the current struct field type []string cannot handle unknown values. Use types.List, or a custom type that supports unknown values instead.",
				),
			},
		},
		"ListType-[]string-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"list": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Map{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Null: true,
			},
		},
		"MapNestedAttributes-types.Map-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Map{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Unknown: true,
			},
		},
		"MapNestedAttributes-types.Map-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Map{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Elems: map[string]attr.Value{
					"key1": types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "value1"},
						},
					},
					"key2": types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "value2"},
						},
					},
				},
			},
		},
		"MapNestedAttributes-map[string]types.Object-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						"Received unknown value for map, however the current struct field type map[string]types.Object cannot handle unknown values. Use types.Map, or a custom type that supports unknown values instead.",
				),
			},
		},
		"MapNestedAttributes-map[string]types.Object-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				"key1": {
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "value1"},
					},
				},
				"key2": {
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "value2"},
					},
				},
			},
		},
		"MapNestedAttributes-map[string]struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						`Received unknown value for map, however the current struct field type map[string]struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Map, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"MapNestedAttributes-map[string]struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				"key1": {NestedString: types.String{Value: "value1"}},
				"key2": {NestedString: types.String{Value: "value2"}},
			},
		},
		"MapType-types.Map-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
			path:   path.Root("map"),
			target: new(types.Map),
			expected: &types.Map{
				ElemType: types.StringType,
				Null:     true,
			},
		},
		"MapType-types.Map-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
			path:   path.Root("map"),
			target: new(types.Map),
			expected: &types.Map{
				ElemType: types.StringType,
				Unknown:  true,
			},
		},
		"MapType-types.Map-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
			expected: &types.Map{
				ElemType: types.StringType,
				Elems: map[string]attr.Value{
					"key1": types.String{Value: "value1"},
					"key2": types.String{Value: "value2"},
				},
			},
		},
		"MapType-map[string]types.String-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
						"Received unknown value for map, however the current struct field type map[string]types.String cannot handle unknown values. Use types.Map, or a custom type that supports unknown values instead.",
				),
			},
		},
		"MapType-map[string]types.String-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
				"key1": {Value: "value1"},
				"key2": {Value: "value2"},
			},
		},
		"MapType-map[string]string-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
						"Received unknown value for map, however the current struct field type map[string]string cannot handle unknown values. Use types.Map, or a custom type that supports unknown values instead.",
				),
			},
		},
		"MapType-map[string]string-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"map": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Null: true,
			},
		},
		"ObjectType-types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Unknown: true,
			},
		},
		"ObjectType-types.Object-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"nested_string": types.String{Value: "test1"},
				},
			},
		},
		"ObjectType-*struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
						`Received unknown value for object, however the current struct field type *struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Object, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"ObjectType-*struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
				NestedString: types.String{Value: "test1"},
			}),
		},
		"ObjectType-struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
						`Received null value for object, however the current struct field type struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle null values. Use a pointer type (*struct { NestedString types.String "tfsdk:\"nested_string\"" }), types.Object, or a custom type that supports null values instead.`,
				),
			},
		},
		"ObjectType-struct-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
						`Received unknown value for object, however the current struct field type struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Object, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"ObjectType-struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
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
				NestedString: types.String{Value: "test1"},
			},
		},
		"SetBlock-types.Set-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
			expected: &types.Set{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Null: true,
			},
		},
		"SetBlock-types.Set-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
			expected: &types.Set{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Unknown: true,
			},
		},
		"SetBlock-types.Set-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
			expected: &types.Set{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "test1"},
						},
					},
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "test2"},
						},
					},
				},
			},
		},
		"SetBlock-[]types.Object-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
						"Received unknown value for set, however the current struct field type []types.Object cannot handle unknown values. Use types.Set, or a custom type that supports unknown values instead.",
				),
			},
		},
		"SetBlock-[]types.Object-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
				{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "test1"},
					},
				},
				{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "test2"},
					},
				},
			},
		},
		"SetBlock-[]struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
						`Received unknown value for set, however the current struct field type []struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Set, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"SetBlock-[]struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"set": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
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
				{NestedString: types.String{Value: "test1"}},
				{NestedString: types.String{Value: "test2"}},
			},
		},
		"SetNestedAttributes-types.Set-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Set{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Null: true,
			},
		},
		"SetNestedAttributes-types.Set-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Set{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Unknown: true,
			},
		},
		"SetNestedAttributes-types.Set-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Set{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "test1"},
						},
					},
					types.Object{
						AttrTypes: map[string]attr.Type{
							"nested_string": types.StringType,
						},
						Attrs: map[string]attr.Value{
							"nested_string": types.String{Value: "test2"},
						},
					},
				},
			},
		},
		"SetNestedAttributes-[]types.Object-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						"Received unknown value for set, however the current struct field type []types.Object cannot handle unknown values. Use types.Set, or a custom type that supports unknown values instead.",
				),
			},
		},
		"SetNestedAttributes-[]types.Object-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "test1"},
					},
				},
				{
					AttrTypes: map[string]attr.Type{
						"nested_string": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"nested_string": types.String{Value: "test2"},
					},
				},
			},
		},
		"SetNestedAttributes-[]struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						`Received unknown value for set, however the current struct field type []struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Set, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"SetNestedAttributes-[]struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				{NestedString: types.String{Value: "test1"}},
				{NestedString: types.String{Value: "test2"}},
			},
		},
		"SetType-types.Set-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: &types.Set{
				ElemType: types.StringType,
				Null:     true,
			},
		},
		"SetType-types.Set-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
			path:   path.Root("set"),
			target: new(types.Set),
			expected: &types.Set{
				ElemType: types.StringType,
				Unknown:  true,
			},
		},
		"SetType-types.Set-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
			expected: &types.Set{
				ElemType: types.StringType,
				Elems: []attr.Value{
					types.String{Value: "test1"},
					types.String{Value: "test2"},
				},
			},
		},
		"SetType-[]types.String-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
						"Received unknown value for set, however the current struct field type []types.String cannot handle unknown values. Use types.Set, or a custom type that supports unknown values instead.",
				),
			},
		},
		"SetType-[]types.String-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
				{Value: "test1"},
				{Value: "test2"},
			},
		},
		"SetType-[]string-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
						"Received unknown value for set, however the current struct field type []string cannot handle unknown values. Use types.Set, or a custom type that supports unknown values instead.",
				),
			},
		},
		"SetType-[]string-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"set": {
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
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Null: true,
			},
		},
		"SingleBlock-types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Unknown: true,
			},
		},
		"SingleBlock-types.Object-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"nested_string": types.String{Value: "test1"},
				},
			},
		},
		"SingleBlock-*struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
						`Received unknown value for object, however the current struct field type *struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Object, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"SingleBlock-*struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
				NestedString: types.String{Value: "test1"},
			}),
		},
		"SingleBlock-struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
						`Received null value for object, however the current struct field type struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle null values. Use a pointer type (*struct { NestedString types.String "tfsdk:\"nested_string\"" }), types.Object, or a custom type that supports null values instead.`,
				),
			},
		},
		"SingleBlock-struct-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
						`Received unknown value for object, however the current struct field type struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Object, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"SingleBlock-struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"object": {
							Attributes: map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
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
				NestedString: types.String{Value: "test1"},
			},
		},
		"SingleNestedAttributes-types.Object-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Null: true,
			},
		},
		"SingleNestedAttributes-types.Object-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Unknown: true,
			},
		},
		"SingleNestedAttributes-types.Object-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
			expected: &types.Object{
				AttrTypes: map[string]attr.Type{
					"nested_string": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"nested_string": types.String{Value: "test1"},
				},
			},
		},
		"SingleNestedAttributes-*struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						`Received unknown value for object, however the current struct field type *struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Object, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"SingleNestedAttributes-*struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				NestedString: types.String{Value: "test1"},
			}),
		},
		"SingleNestedAttributes-struct-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						`Received null value for object, however the current struct field type struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle null values. Use a pointer type (*struct { NestedString types.String "tfsdk:\"nested_string\"" }), types.Object, or a custom type that supports null values instead.`,
				),
			},
		},
		"SingleNestedAttributes-struct-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
						`Received unknown value for object, however the current struct field type struct { NestedString types.String "tfsdk:\"nested_string\"" } cannot handle unknown values. Use types.Object, or a custom type that supports unknown values instead.`,
				),
			},
		},
		"SingleNestedAttributes-struct-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"object": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"nested_string": {
									Optional: true,
									Type:     types.StringType,
								},
							}),
							Optional: true,
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
				NestedString: types.String{Value: "test1"},
			},
		},
		"StringType-types.string-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
			expected: &types.String{Null: true},
		},
		"StringType-types.string-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
			expected: &types.String{Unknown: true},
		},
		"StringType-types.string-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
			expected: &types.String{Value: "test"},
		},
		"StringType-*string-null": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
						"Received unknown value for string, however the current struct field type *string cannot handle unknown values. Use types.String, or a custom type that supports unknown values instead.",
				),
			},
		},
		"StringType-*string-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
						`Received null value for string, however the current struct field type string cannot handle null values. Use a pointer type (*string), types.String, or a custom type that supports null values instead.`,
				),
			},
		},
		"StringType-string-unknown": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
						"Received unknown value for string, however the current struct field type string cannot handle unknown values. Use types.String, or a custom type that supports unknown values instead.",
				),
			},
		},
		"StringType-string-value": {
			data: fwschemadata.Data{
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"string": {
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
