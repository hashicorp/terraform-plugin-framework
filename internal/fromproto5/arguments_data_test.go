// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestArgumentsData(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input             []*tfprotov5.DynamicValue
		definition        function.Definition
		expected          function.ArgumentsData
		expectedFuncError *function.FuncError
		expectedLog       []map[string]interface{}
	}{
		"nil": {
			input:      nil,
			definition: function.Definition{},
			expected:   function.ArgumentsData{},
		},
		"empty": {
			input:      []*tfprotov5.DynamicValue{},
			definition: function.Definition{},
			expected:   function.ArgumentsData{},
		},
		"mismatched-arguments-too-few-arguments": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, nil)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.BoolParameter{},
				},
			},
			expected: function.ArgumentsData{},
			expectedFuncError: function.NewFuncError(
				"Unexpected Function Arguments Data: " +
					"The provider received an unexpected number of function arguments from Terraform for the given function definition. " +
					"This is always an issue in terraform-plugin-framework or Terraform itself and should be reported to the provider developers.\n\n" +
					"Expected function arguments: 2\n" +
					"Given function arguments: 1",
			),
		},
		"mismatched-arguments-too-many-arguments": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, nil)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, nil)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
			},
			expected: function.ArgumentsData{},
			expectedFuncError: function.NewFuncError(
				"Unexpected Function Arguments Data: " +
					"The provider received an unexpected number of function arguments from Terraform for the given function definition. " +
					"This is always an issue in terraform-plugin-framework or Terraform itself and should be reported to the provider developers.\n\n" +
					"Expected function arguments: 1\n" +
					"Given function arguments: 2",
			),
		},
		"mismatched-arguments-type": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{},
				},
			},
			expected: function.ArgumentsData{},
			expectedFuncError: function.NewArgumentFuncError(
				0,
				"Unable to Convert Function Argument: "+
					"An unexpected error was encountered when converting the function argument from the protocol type. "+
					"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
					"Please report this to the provider developer:\n\n"+
					"Unable to unmarshal DynamicValue at position 0: error decoding string: msgpack: invalid code=c3 decoding string/bytes length",
			),
		},
		"parameters-zero": {
			input:      []*tfprotov5.DynamicValue{},
			definition: function.Definition{},
			expected:   function.NewArgumentsData(nil),
		},
		"parameters-one": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
		},
		"parameters-one-CustomType": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						CustomType: testtypes.BoolType{},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				testtypes.Bool{
					Bool: basetypes.NewBoolValue(true),
				},
			}),
		},
		"parameters-one-TypeWithValidation-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateError{},
					},
				},
			},
			expected: function.NewArgumentsData(nil),
			expectedFuncError: function.NewArgumentFuncError(
				0,
				"Error Diagnostic: This is an error.",
			),
		},
		"parameters-one-TypeWithParameterValidation-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateParameterError{},
					},
				},
			},
			expected: function.NewArgumentsData(nil),
			expectedFuncError: function.NewArgumentFuncError(
				0,
				"This is a function error",
			),
		},
		"parameters-one-TypeWithValidation-warning": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateWarning{},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				testtypes.Bool{
					Bool: basetypes.NewBoolValue(true),
				},
			}),
			expectedLog: []map[string]interface{}{
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "This is a warning.",
					"summary":  "Warning Diagnostic",
				},
			},
		},
		"parameters-one-variadicparameter-zero": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewTupleValueMust(
					[]attr.Type{},
					[]attr.Value{},
				),
			}),
		},
		"parameters-one-variadicparameter-one": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg1")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg1"),
					},
				),
			}),
		},
		"parameters-one-variadicparameter-multiple": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg1")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg2")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg1"),
						basetypes.NewStringValue("varg-arg2"),
					},
				),
			}),
		},
		"parameters-multiple": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, false)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.BoolParameter{},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewBoolValue(false),
			}),
		},
		"parameters-multiple-TypeWithValidation-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, false)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateError{},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				1,
				"Error Diagnostic: This is an error.",
			),
		},
		"parameters-multiple-TypeWithParameterValidation-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, false)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateParameterError{},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				1,
				"This is a function error",
			),
		},
		"parameters-multiple-TypeWithValidation-warning": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, false)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateWarning{},
					},
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateWarning{},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				testtypes.Bool{
					Bool:      basetypes.NewBoolValue(true),
					CreatedBy: testtypes.BoolTypeWithValidateWarning{},
				},
				testtypes.Bool{
					Bool:      basetypes.NewBoolValue(false),
					CreatedBy: testtypes.BoolTypeWithValidateWarning{},
				},
			}),
			expectedLog: []map[string]interface{}{
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "This is a warning.",
					"summary":  "Warning Diagnostic",
				},
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "This is a warning.",
					"summary":  "Warning Diagnostic",
				},
			},
		},
		"parameters-multiple-TypeWithValidation-warning-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, false)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateWarning{},
					},
					function.BoolParameter{
						CustomType: testtypes.BoolTypeWithValidateError{},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				testtypes.Bool{
					Bool:      basetypes.NewBoolValue(true),
					CreatedBy: testtypes.BoolTypeWithValidateWarning{},
				},
			}),
			expectedFuncError: function.NewArgumentFuncError(
				1,
				"Error Diagnostic: This is an error.",
			),
			expectedLog: []map[string]interface{}{
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "This is a warning.",
					"summary":  "Warning Diagnostic",
				},
			},
		},
		"parameters-multiple-variadicparameter-zero": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, false)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewBoolValue(false),
				basetypes.NewTupleValueMust(
					[]attr.Type{},
					[]attr.Value{},
				),
			}),
		},
		"parameters-multiple-variadicparameter-one": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, false)),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg2")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewBoolValue(false),
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg2"),
					},
				),
			}),
		},
		"parameters-multiple-variadicparameter-multiple": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, false)),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg2")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg3")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewBoolValue(false),
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg2"),
						basetypes.NewStringValue("varg-arg3"),
					},
				),
			}),
		},
		"variadicparameter-zero": {
			input: []*tfprotov5.DynamicValue{},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{},
					[]attr.Value{},
				),
			}),
		},
		"variadicparameter-one": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg0"),
					},
				),
			}),
		},
		"variadicparameter-one-CustomType": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					CustomType: testtypes.StringType{},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						testtypes.StringType{},
					},
					[]attr.Value{
						testtypes.String{
							CreatedBy:      testtypes.StringType{},
							InternalString: basetypes.NewStringValue("varg-arg0"),
						},
					},
				),
			}),
		},
		"variadicparameter-one-TypeWithValidation-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					CustomType: testtypes.StringTypeWithValidateError{},
				},
			},
			expected: function.NewArgumentsData(nil),
			expectedFuncError: function.NewArgumentFuncError(
				0,
				"Error Diagnostic: This is an error.",
			),
		},
		"variadicparameter-one-TypeWithParameterValidation-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					CustomType: testtypes.StringTypeWithValidateParameterError{},
				},
			},
			expected: function.NewArgumentsData(nil),
			expectedFuncError: function.NewArgumentFuncError(
				0,
				"This is a function error",
			),
		},
		"variadicparameter-one-TypeWithValidation-warning": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					CustomType: testtypes.StringTypeWithValidateWarning{},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						testtypes.StringTypeWithValidateWarning{},
					},
					[]attr.Value{
						testtypes.String{
							CreatedBy:      testtypes.StringTypeWithValidateWarning{},
							InternalString: basetypes.NewStringValue("varg-arg0"),
						},
					},
				),
			}),
			expectedLog: []map[string]interface{}{
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "This is a warning.",
					"summary":  "Warning Diagnostic",
				},
			},
		},
		"variadicparameter-multiple": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg1")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg0"),
						basetypes.NewStringValue("varg-arg1"),
					},
				),
			}),
		},
		"variadicparameter-multiple-TypeWithValidation-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg1")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					CustomType: testtypes.StringTypeWithValidateError{},
				},
			},
			expected: function.NewArgumentsData(nil),
			expectedFuncError: function.NewArgumentFuncError(
				0,
				"Error Diagnostic: This is an error.\nError Diagnostic: This is an error.",
			),
		},
		"variadicparameter-multiple-TypeWithParameterValidation-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg1")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					CustomType: testtypes.StringTypeWithValidateParameterError{},
				},
			},
			expected: function.NewArgumentsData(nil),
			expectedFuncError: function.NewArgumentFuncError(
				0,
				"This is a function error\nThis is a function error",
			),
		},
		"variadicparameter-multiple-TypeWithValidation-warning": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg1")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					CustomType: testtypes.StringTypeWithValidateWarning{},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						testtypes.StringTypeWithValidateWarning{},
						testtypes.StringTypeWithValidateWarning{},
					},
					[]attr.Value{
						testtypes.String{
							CreatedBy:      testtypes.StringTypeWithValidateWarning{},
							InternalString: basetypes.NewStringValue("varg-arg0"),
						},
						testtypes.String{
							CreatedBy:      testtypes.StringTypeWithValidateWarning{},
							InternalString: basetypes.NewStringValue("varg-arg1"),
						},
					},
				),
			}),
			expectedLog: []map[string]interface{}{
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "This is a warning.",
					"summary":  "Warning Diagnostic",
				},
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "This is a warning.",
					"summary":  "Warning Diagnostic",
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			ctx := tflogtest.RootLogger(context.Background(), &output)

			got, diags := fromproto5.ArgumentsData(ctx, testCase.input, testCase.definition)

			entries, err := tflogtest.MultilineJSONDecode(&output)

			if err != nil {
				t.Fatalf("unable to read multiple line JSON: %s", err)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedFuncError); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(entries, testCase.expectedLog); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
