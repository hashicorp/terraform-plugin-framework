// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestArgumentsData(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input               []*tfprotov6.DynamicValue
		definition          function.Definition
		expected            function.ArgumentsData
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:      nil,
			definition: function.Definition{},
			expected:   function.ArgumentsData{},
		},
		"empty": {
			input:      []*tfprotov6.DynamicValue{},
			definition: function.Definition{},
			expected:   function.ArgumentsData{},
		},
		"mismatched-arguments-too-few-arguments": {
			input: []*tfprotov6.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, nil)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.BoolParameter{},
				},
			},
			expected: function.ArgumentsData{},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unexpected Function Arguments Data",
					"The provider received an unexpected number of function arguments from Terraform for the given function definition. "+
						"This is always an issue in terraform-plugin-framework or Terraform itself and should be reported to the provider developers.\n\n"+
						"Expected function arguments: 2\n"+
						"Given function arguments: 1",
				),
			},
		},
		"mismatched-arguments-too-many-arguments": {
			input: []*tfprotov6.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, nil)),
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, nil)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
			},
			expected: function.ArgumentsData{},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unexpected Function Arguments Data",
					"The provider received an unexpected number of function arguments from Terraform for the given function definition. "+
						"This is always an issue in terraform-plugin-framework or Terraform itself and should be reported to the provider developers.\n\n"+
						"Expected function arguments: 1\n"+
						"Given function arguments: 2",
				),
			},
		},
		"mismatched-arguments-type": {
			input: []*tfprotov6.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			},
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{},
				},
			},
			expected: function.ArgumentsData{},
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Function Argument",
					"An unexpected error was encountered when converting the function argument from the protocol type. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Unable to unmarshal DynamicValue at position 0: error decoding string: msgpack: invalid code=c3 decoding string/bytes length",
				),
			},
		},
		"parameters-zero": {
			input:      []*tfprotov6.DynamicValue{},
			definition: function.Definition{},
			expected:   function.NewArgumentsData(nil),
		},
		"parameters-one": {
			input: []*tfprotov6.DynamicValue{
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
		"parameters-one-variadicparameter-zero": {
			input: []*tfprotov6.DynamicValue{
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
				basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{},
				),
			}),
		},
		"parameters-one-variadicparameter-one": {
			input: []*tfprotov6.DynamicValue{
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
				basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg1"),
					},
				),
			}),
		},
		"parameters-one-variadicparameter-multiple": {
			input: []*tfprotov6.DynamicValue{
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
				basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg1"),
						basetypes.NewStringValue("varg-arg2"),
					},
				),
			}),
		},
		"parameters-multiple": {
			input: []*tfprotov6.DynamicValue{
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
			input: []*tfprotov6.DynamicValue{
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
				basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{},
				),
			}),
		},
		"parameters-multiple-variadicparameter-one": {
			input: []*tfprotov6.DynamicValue{
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
				basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg2"),
					},
				),
			}),
		},
		"parameters-multiple-variadicparameter-multiple": {
			input: []*tfprotov6.DynamicValue{
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
				basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg2"),
						basetypes.NewStringValue("varg-arg3"),
					},
				),
			}),
		},
		"variadicparameter-zero": {
			input: []*tfprotov6.DynamicValue{},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{},
				),
			}),
		},
		"variadicparameter-one": {
			input: []*tfprotov6.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{
						basetypes.NewStringValue("varg-arg0"),
					},
				),
			}),
		},
		"variadicparameter-multiple": {
			input: []*tfprotov6.DynamicValue{
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg0")),
				DynamicValueMust(tftypes.NewValue(tftypes.String, "varg-arg1")),
			},
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
			},
			expected: function.NewArgumentsData([]attr.Value{
				basetypes.NewListValueMust(
					basetypes.StringType{},
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

			got, diags := fromproto6.ArgumentsData(context.Background(), testCase.input, testCase.definition)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
