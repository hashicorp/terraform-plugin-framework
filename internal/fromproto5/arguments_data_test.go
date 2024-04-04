// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestArgumentsData(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input             []*tfprotov5.DynamicValue
		definition        function.Definition
		expected          function.ArgumentsData
		expectedFuncError *function.FuncError
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
				0, "Unable to Convert Function Argument: "+
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
				0, "Error Diagnostic: This is an error.",
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
				0, "This is a function error",
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
			// Function error is not generated as diagnostic raised is warning.
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.ArgumentsData(context.Background(), testCase.input, testCase.definition)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedFuncError); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestArgumentsData_ParameterValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input             []*tfprotov5.DynamicValue
		definition        function.Definition
		expected          function.ArgumentsData
		expectedFuncError *function.FuncError
	}{
		"bool-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						Validators: []function.BoolValidator{
							testvalidator.Bool{
								ValidateMethod: func(ctx context.Context, req function.BoolRequest, resp *function.BoolResponse) {
									got := req.Value
									expected := types.BoolValue(true)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
		},
		"bool-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						Validators: []function.BoolValidator{
							testvalidator.Bool{
								ValidateMethod: func(ctx context.Context, req function.BoolRequest, resp *function.BoolResponse) {
									got := req.Value
									expected := types.BoolValue(false)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"bool-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						Validators: []function.BoolValidator{
							testvalidator.Bool{
								ValidateMethod: func(ctx context.Context, req function.BoolRequest, resp *function.BoolResponse) {
									got := req.Value
									expected := types.BoolValue(false)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.Bool{
								ValidateMethod: func(ctx context.Context, req function.BoolRequest, resp *function.BoolResponse) {
									got := req.Value
									expected := types.BoolValue(false)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"dynamic-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				createDynamicValue(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.DynamicParameter{
						Validators: []function.DynamicValidator{
							testvalidator.Dynamic{
								ValidateMethod: func(ctx context.Context, req function.DynamicRequest, resp *function.DynamicResponse) {
									got := req.Value
									expected := types.DynamicValue(types.BoolValue(true))

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewDynamicValue(types.BoolValue(true)),
			}),
		},
		"dynamic-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				createDynamicValue(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.DynamicParameter{
						Validators: []function.DynamicValidator{
							testvalidator.Dynamic{
								ValidateMethod: func(ctx context.Context, req function.DynamicRequest, resp *function.DynamicResponse) {
									got := req.Value
									expected := types.DynamicValue(types.BoolValue(false))

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewDynamicValue(types.BoolValue(true)),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"dynamic-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				createDynamicValue(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.DynamicParameter{
						Validators: []function.DynamicValidator{
							testvalidator.Dynamic{
								ValidateMethod: func(ctx context.Context, req function.DynamicRequest, resp *function.DynamicResponse) {
									got := req.Value
									expected := types.DynamicValue(types.BoolValue(false))

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.Dynamic{
								ValidateMethod: func(ctx context.Context, req function.DynamicRequest, resp *function.DynamicResponse) {
									got := req.Value
									expected := types.DynamicValue(types.BoolValue(false))

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewDynamicValue(types.BoolValue(true)),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"float64-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1.0)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.Float64Parameter{
						Validators: []function.Float64Validator{
							testvalidator.Float64{
								ValidateMethod: func(ctx context.Context, req function.Float64Request, resp *function.Float64Response) {
									got := req.Value
									expected := types.Float64Value(1.0)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewFloat64Value(1.0),
			}),
		},
		"float64-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1.0)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.Float64Parameter{
						Validators: []function.Float64Validator{
							testvalidator.Float64{
								ValidateMethod: func(ctx context.Context, req function.Float64Request, resp *function.Float64Response) {
									got := req.Value
									expected := types.Float64Value(2.0)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewFloat64Value(1.0),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"float64-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1.0)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.Float64Parameter{
						Validators: []function.Float64Validator{
							testvalidator.Float64{
								ValidateMethod: func(ctx context.Context, req function.Float64Request, resp *function.Float64Response) {
									got := req.Value
									expected := types.Float64Value(2.0)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.Float64{
								ValidateMethod: func(ctx context.Context, req function.Float64Request, resp *function.Float64Response) {
									got := req.Value
									expected := types.Float64Value(3.0)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewFloat64Value(1.0),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"int64-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.Int64Parameter{
						Validators: []function.Int64Validator{
							testvalidator.Int64{
								ValidateMethod: func(ctx context.Context, req function.Int64Request, resp *function.Int64Response) {
									got := req.Value
									expected := types.Int64Value(1)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewInt64Value(1),
			}),
		},
		"int64-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.Int64Parameter{
						Validators: []function.Int64Validator{
							testvalidator.Int64{
								ValidateMethod: func(ctx context.Context, req function.Int64Request, resp *function.Int64Response) {
									got := req.Value
									expected := types.Int64Value(2)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewInt64Value(1),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"int64-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.Int64Parameter{
						Validators: []function.Int64Validator{
							testvalidator.Int64{
								ValidateMethod: func(ctx context.Context, req function.Int64Request, resp *function.Int64Response) {
									got := req.Value
									expected := types.Int64Value(2)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.Int64{
								ValidateMethod: func(ctx context.Context, req function.Int64Request, resp *function.Int64Response) {
									got := req.Value
									expected := types.Int64Value(3)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewInt64Value(1),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"list-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.List{ElementType: tftypes.Bool}, []tftypes.Value{tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.ListParameter{
						ElementType: types.BoolType,
						Validators: []function.ListValidator{
							testvalidator.List{
								ValidateMethod: func(ctx context.Context, req function.ListRequest, resp *function.ListResponse) {
									got := req.Value
									expected, _ := types.ListValue(types.BoolType, []attr.Value{types.BoolValue(true)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createListValue(types.BoolType, []attr.Value{types.BoolValue(true)}),
			}),
		},
		"list-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.List{ElementType: tftypes.Bool}, []tftypes.Value{tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.ListParameter{
						ElementType: types.BoolType,
						Validators: []function.ListValidator{
							testvalidator.List{
								ValidateMethod: func(ctx context.Context, req function.ListRequest, resp *function.ListResponse) {
									got := req.Value
									expected, _ := types.ListValue(types.BoolType, []attr.Value{types.BoolValue(true),
										types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createListValue(types.BoolType, []attr.Value{types.BoolValue(true)}),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"list-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.List{ElementType: tftypes.Bool}, []tftypes.Value{tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.ListParameter{
						ElementType: types.BoolType,
						Validators: []function.ListValidator{
							testvalidator.List{
								ValidateMethod: func(ctx context.Context, req function.ListRequest, resp *function.ListResponse) {
									got := req.Value
									expected, _ := types.ListValue(types.BoolType, []attr.Value{types.BoolValue(true),
										types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.List{
								ValidateMethod: func(ctx context.Context, req function.ListRequest, resp *function.ListResponse) {
									got := req.Value
									expected, _ := types.ListValue(types.BoolType, []attr.Value{types.BoolValue(true),
										types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createListValue(types.BoolType, []attr.Value{types.BoolValue(true)}),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"map-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Map{ElementType: tftypes.Bool},
					map[string]tftypes.Value{"key": tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.MapParameter{
						ElementType: types.BoolType,
						Validators: []function.MapValidator{
							testvalidator.Map{
								ValidateMethod: func(ctx context.Context, req function.MapRequest, resp *function.MapResponse) {
									got := req.Value
									expected, _ := types.MapValue(types.BoolType, map[string]attr.Value{"key": types.BoolValue(true)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createMapValue(types.BoolType, map[string]attr.Value{"key": types.BoolValue(true)}),
			}),
		},
		"map-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Map{ElementType: tftypes.Bool},
					map[string]tftypes.Value{"key": tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.MapParameter{
						ElementType: types.BoolType,
						Validators: []function.MapValidator{
							testvalidator.Map{
								ValidateMethod: func(ctx context.Context, req function.MapRequest, resp *function.MapResponse) {
									got := req.Value
									expected, _ := types.MapValue(types.BoolType, map[string]attr.Value{"key": types.BoolValue(true),
										"key2": types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createMapValue(types.BoolType, map[string]attr.Value{"key": types.BoolValue(true)}),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"map-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Map{ElementType: tftypes.Bool},
					map[string]tftypes.Value{"key": tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.MapParameter{
						ElementType: types.BoolType,
						Validators: []function.MapValidator{
							testvalidator.Map{
								ValidateMethod: func(ctx context.Context, req function.MapRequest, resp *function.MapResponse) {
									got := req.Value
									expected, _ := types.MapValue(types.BoolType, map[string]attr.Value{"key": types.BoolValue(true),
										"key2": types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.Map{
								ValidateMethod: func(ctx context.Context, req function.MapRequest, resp *function.MapResponse) {
									got := req.Value
									expected, _ := types.MapValue(types.BoolType, map[string]attr.Value{"key1": types.BoolValue(true),
										"key2": types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createMapValue(types.BoolType, map[string]attr.Value{"key": types.BoolValue(true)}),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"number-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.NumberParameter{
						Validators: []function.NumberValidator{
							testvalidator.Number{
								ValidateMethod: func(ctx context.Context, req function.NumberRequest, resp *function.NumberResponse) {
									got := req.Value
									expected := types.NumberValue(big.NewFloat(1))

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewNumberValue(big.NewFloat(1)),
			}),
		},
		"number-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.NumberParameter{
						Validators: []function.NumberValidator{
							testvalidator.Number{
								ValidateMethod: func(ctx context.Context, req function.NumberRequest, resp *function.NumberResponse) {
									got := req.Value
									expected := types.NumberValue(big.NewFloat(2))

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewNumberValue(big.NewFloat(1)),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"number-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Number, 1)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.NumberParameter{
						Validators: []function.NumberValidator{
							testvalidator.Number{
								ValidateMethod: func(ctx context.Context, req function.NumberRequest, resp *function.NumberResponse) {
									got := req.Value
									expected := types.NumberValue(big.NewFloat(2))

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.Number{
								ValidateMethod: func(ctx context.Context, req function.NumberRequest, resp *function.NumberResponse) {
									got := req.Value
									expected := types.NumberValue(big.NewFloat(3))

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewNumberValue(big.NewFloat(1)),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"object-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"boolAttribute": tftypes.Bool}},
					map[string]tftypes.Value{"boolAttribute": tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.ObjectParameter{
						AttributeTypes: map[string]attr.Type{
							"boolAttribute": types.BoolType,
						},
						Validators: []function.ObjectValidator{
							testvalidator.Object{
								ValidateMethod: func(ctx context.Context, req function.ObjectRequest, resp *function.ObjectResponse) {
									got := req.Value
									expected, _ := types.ObjectValue(map[string]attr.Type{"boolAttribute": types.BoolType},
										map[string]attr.Value{"boolAttribute": types.BoolValue(true)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createObjectValue(map[string]attr.Type{"boolAttribute": types.BoolType},
					map[string]attr.Value{"boolAttribute": types.BoolValue(true)}),
			}),
		},
		"object-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"boolAttribute": tftypes.Bool}},
					map[string]tftypes.Value{"boolAttribute": tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.ObjectParameter{
						AttributeTypes: map[string]attr.Type{
							"boolAttribute": types.BoolType,
						},
						Validators: []function.ObjectValidator{
							testvalidator.Object{
								ValidateMethod: func(ctx context.Context, req function.ObjectRequest, resp *function.ObjectResponse) {
									got := req.Value
									expected, _ := types.ObjectValue(map[string]attr.Type{"boolAttribute": types.BoolType,
										"boolAttribute2": types.BoolType},
										map[string]attr.Value{"boolAttribute": types.BoolValue(true),
											"boolAttribute2": types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createObjectValue(map[string]attr.Type{"boolAttribute": types.BoolType},
					map[string]attr.Value{"boolAttribute": types.BoolValue(true)}),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"object-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"boolAttribute": tftypes.Bool}},
					map[string]tftypes.Value{"boolAttribute": tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.ObjectParameter{
						AttributeTypes: map[string]attr.Type{
							"boolAttribute": types.BoolType,
						},
						Validators: []function.ObjectValidator{
							testvalidator.Object{
								ValidateMethod: func(ctx context.Context, req function.ObjectRequest, resp *function.ObjectResponse) {
									got := req.Value
									expected, _ := types.ObjectValue(map[string]attr.Type{"boolAttribute": types.BoolType,
										"boolAttribute2": types.BoolType},
										map[string]attr.Value{"boolAttribute": types.BoolValue(true),
											"boolAttribute2": types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.Object{
								ValidateMethod: func(ctx context.Context, req function.ObjectRequest, resp *function.ObjectResponse) {
									got := req.Value
									expected, _ := types.ObjectValue(map[string]attr.Type{"boolAttribute1": types.BoolType,
										"boolAttribute2": types.BoolType},
										map[string]attr.Value{"boolAttribute1": types.BoolValue(true),
											"boolAttribute2": types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createObjectValue(map[string]attr.Type{"boolAttribute": types.BoolType},
					map[string]attr.Value{"boolAttribute": types.BoolValue(true)}),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"set-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Set{ElementType: tftypes.Bool}, []tftypes.Value{tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.SetParameter{
						ElementType: types.BoolType,
						Validators: []function.SetValidator{
							testvalidator.Set{
								ValidateMethod: func(ctx context.Context, req function.SetRequest, resp *function.SetResponse) {
									got := req.Value
									expected, _ := types.SetValue(types.BoolType, []attr.Value{types.BoolValue(true)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createSetValue(types.BoolType, []attr.Value{types.BoolValue(true)}),
			}),
		},
		"set-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Set{ElementType: tftypes.Bool}, []tftypes.Value{tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.SetParameter{
						ElementType: types.BoolType,
						Validators: []function.SetValidator{
							testvalidator.Set{
								ValidateMethod: func(ctx context.Context, req function.SetRequest, resp *function.SetResponse) {
									got := req.Value
									expected, _ := types.SetValue(types.BoolType, []attr.Value{types.BoolValue(true),
										types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createSetValue(types.BoolType, []attr.Value{types.BoolValue(true)}),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"set-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Set{ElementType: tftypes.Bool}, []tftypes.Value{tftypes.NewValue(tftypes.Bool, true)})),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.SetParameter{
						ElementType: types.BoolType,
						Validators: []function.SetValidator{
							testvalidator.Set{
								ValidateMethod: func(ctx context.Context, req function.SetRequest, resp *function.SetResponse) {
									got := req.Value
									expected, _ := types.SetValue(types.BoolType, []attr.Value{types.BoolValue(true),
										types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.Set{
								ValidateMethod: func(ctx context.Context, req function.SetRequest, resp *function.SetResponse) {
									got := req.Value
									expected, _ := types.SetValue(types.BoolType, []attr.Value{types.BoolValue(true),
										types.BoolValue(false)})

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				createSetValue(types.BoolType, []attr.Value{types.BoolValue(true)}),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"string-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "true")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{
						Validators: []function.StringValidator{
							testvalidator.String{
								ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
									got := req.Value
									expected := types.StringValue("true")

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewStringValue("true"),
			}),
		},
		"string-parameter-Validators-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "true")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{
						Validators: []function.StringValidator{
							testvalidator.String{
								ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
									got := req.Value
									expected := types.StringValue("false")

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewStringValue("true"),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0, "Error Diagnostic: This is an error.",
			),
		},
		"string-parameter-Validators-multiple-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "true")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{
						Validators: []function.StringValidator{
							testvalidator.String{
								ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
									got := req.Value
									expected := types.StringValue("false")

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 1.",
										)
									}
								},
							},
							testvalidator.String{
								ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
									got := req.Value
									expected := types.StringValue("false")

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: error 2.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewStringValue("true"),
			}),
			expectedFuncError: function.NewArgumentFuncError(
				0,
				"Error Diagnostic: error 1."+
					"\nError Diagnostic: error 2.",
			),
		},
		"multiple-parameter-Validators": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "true")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						Validators: []function.BoolValidator{
							testvalidator.Bool{
								ValidateMethod: func(ctx context.Context, req function.BoolRequest, resp *function.BoolResponse) {
									got := req.Value
									expected := types.BoolValue(true)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
					function.StringParameter{
						Validators: []function.StringValidator{
							testvalidator.String{
								ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
									got := req.Value
									expected := types.StringValue("true")

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewStringValue("true"),
			}),
		},
		"multiple-parameter-Validators-errors": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "true")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						Validators: []function.BoolValidator{
							testvalidator.Bool{
								ValidateMethod: func(ctx context.Context, req function.BoolRequest, resp *function.BoolResponse) {
									got := req.Value
									expected := types.BoolValue(false)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: bool validator error.",
										)
									}
								},
							},
						},
					},
					function.StringParameter{
						Validators: []function.StringValidator{
							testvalidator.String{
								ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
									got := req.Value
									expected := types.StringValue("false")

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: string validator error.",
										)
									}
								},
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewStringValue("true"),
			}),
			expectedFuncError: function.ConcatFuncErrors(
				function.NewArgumentFuncError(0, "Error Diagnostic: bool validator error."),
				function.NewArgumentFuncError(1, "Error Diagnostic: string validator error."),
			),
		},
		"variadicparameter-one": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "false")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					Validators: []function.StringValidator{
						testvalidator.String{
							ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
								got := req.Value
								expected := types.StringValue("false")

								if !got.Equal(expected) {
									resp.Error = function.NewArgumentFuncError(
										req.ArgumentPosition,
										"Error Diagnostic: string validator error.",
									)
								}
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("false"),
					},
				),
			}),
		},
		"variadicparameter-one-error": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "false")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					Validators: []function.StringValidator{
						testvalidator.String{
							ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
								got := req.Value
								expected := types.StringValue("true")

								if !got.Equal(expected) {
									resp.Error = function.NewArgumentFuncError(
										req.ArgumentPosition,
										"Error Diagnostic: string validator error.",
									)
								}
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("false"),
					},
				),
			}),
			expectedFuncError: function.ConcatFuncErrors(
				function.NewArgumentFuncError(0, "Error Diagnostic: string validator error."),
			),
		},
		"variadicparameter-multiple": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "false")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "false")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					Validators: []function.StringValidator{
						testvalidator.String{
							ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
								got := req.Value
								expected := types.StringValue("false")

								if !got.Equal(expected) {
									resp.Error = function.NewArgumentFuncError(
										req.ArgumentPosition,
										"Error Diagnostic: string validator error.",
									)
								}
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("false"),
						basetypes.NewStringValue("false"),
					},
				),
			}),
		},
		"variadicparameter-multiple-error-single": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "true")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "false")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					Validators: []function.StringValidator{
						testvalidator.String{
							ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
								got := req.Value
								expected := types.StringValue("true")

								if !got.Equal(expected) {
									resp.Error = function.NewArgumentFuncError(
										req.ArgumentPosition,
										"Error Diagnostic: string validator error.",
									)
								}
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("true"),
						basetypes.NewStringValue("false"),
					},
				),
			}),
			expectedFuncError: function.ConcatFuncErrors(
				function.NewArgumentFuncError(1, "Error Diagnostic: string validator error."),
			),
		},
		"variadicparameter-multiple-errors-multiple": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "false")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "false")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{
					Validators: []function.StringValidator{
						testvalidator.String{
							ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
								got := req.Value
								expected := types.StringValue("true")

								if !got.Equal(expected) {
									resp.Error = function.NewArgumentFuncError(
										req.ArgumentPosition,
										"Error Diagnostic: string validator error.",
									)
								}
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("false"),
						basetypes.NewStringValue("false"),
					},
				),
			}),
			expectedFuncError: function.ConcatFuncErrors(
				function.NewArgumentFuncError(0, "Error Diagnostic: string validator error."),
				function.NewArgumentFuncError(0, "Error Diagnostic: string validator error."),
			),
		},
		"boolparameter-and-variadicparameter-multiple-error-single": {
			input: []*tfprotov5.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "true")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "false")),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						Validators: []function.BoolValidator{
							testvalidator.Bool{
								ValidateMethod: func(ctx context.Context, req function.BoolRequest, resp *function.BoolResponse) {
									got := req.Value
									expected := types.BoolValue(true)

									if !got.Equal(expected) {
										resp.Error = function.NewArgumentFuncError(
											req.ArgumentPosition,
											"Error Diagnostic: This is an error.",
										)
									}
								},
							},
						},
					},
				},
				VariadicParameter: function.StringParameter{
					Validators: []function.StringValidator{
						testvalidator.String{
							ValidateMethod: func(ctx context.Context, req function.StringRequest, resp *function.StringResponse) {
								got := req.Value
								expected := types.StringValue("true")

								if !got.Equal(expected) {
									resp.Error = function.NewArgumentFuncError(
										req.ArgumentPosition,
										"Error Diagnostic: string validator error.",
									)
								}
							},
						},
					},
				},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("true"),
						basetypes.NewStringValue("false"),
					},
				),
			}),
			expectedFuncError: function.ConcatFuncErrors(
				function.NewArgumentFuncError(2, "Error Diagnostic: string validator error."),
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto5.ArgumentsData(context.Background(), testCase.input, testCase.definition)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedFuncError); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func createListValue(elementType attr.Type, elements []attr.Value) attr.Value {
	list, _ := basetypes.NewListValue(elementType, elements)
	return list
}

func createMapValue(elementType attr.Type, elements map[string]attr.Value) attr.Value {
	mapVal, _ := basetypes.NewMapValue(elementType, elements)
	return mapVal
}

func createObjectValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) attr.Value {
	object, _ := basetypes.NewObjectValue(attributeTypes, attributes)
	return object
}

func createSetValue(elementType attr.Type, elements []attr.Value) attr.Value {
	list, _ := basetypes.NewSetValue(elementType, elements)
	return list
}

func createDynamicValue(value tftypes.Value) *tfprotov5.DynamicValue {
	dynamicVal, _ := tfprotov5.NewDynamicValue(tftypes.DynamicPseudoType, value)
	return &dynamicVal
}
