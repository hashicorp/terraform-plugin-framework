// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

func TestDefinitionParameter(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		definition          function.Definition
		position            int
		expected            function.Parameter
		expectedDiagnostics diag.Diagnostics
	}{
		"none": {
			definition: function.Definition{
				// no Parameters or VariadicParameter
			},
			position: 0,
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Parameter Position for Definition",
					"When determining the parameter for the given argument position, an invalid value was given. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Function does not implement parameters.\n"+
						"Given position: 0",
				),
			},
		},
		"parameters-first": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.Int64Parameter{},
					function.StringParameter{},
				},
			},
			position: 0,
			expected: function.BoolParameter{},
		},
		"parameters-last": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.Int64Parameter{},
					function.StringParameter{},
				},
			},
			position: 2,
			expected: function.StringParameter{},
		},
		"parameters-middle": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.Int64Parameter{},
					function.StringParameter{},
				},
			},
			position: 1,
			expected: function.Int64Parameter{},
		},
		"parameters-only": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
			},
			position: 0,
			expected: function.BoolParameter{},
		},
		"parameters-over": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
			},
			position: 1,
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Parameter Position for Definition",
					"When determining the parameter for the given argument position, an invalid value was given. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Max argument position: 0\n"+
						"Given position: 1",
				),
			},
		},
		"variadicparameter-and-parameters-select-parameter": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			position: 0,
			expected: function.BoolParameter{},
		},
		"variadicparameter-and-parameters-select-variadicparameter": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
				},
				VariadicParameter: function.StringParameter{},
			},
			position: 1,
			expected: function.StringParameter{},
		},
		"variadicparameter-only": {
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
			},
			position: 0,
			expected: function.StringParameter{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := testCase.definition.Parameter(context.Background(), testCase.position)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestDefinitionValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		definition function.Definition
		expected   diag.Diagnostics
	}{
		"valid": {
			definition: function.Definition{
				Return: function.StringReturn{},
			},
		},
		"result-missing": {
			definition: function.Definition{},
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Function Definition",
					"When validating the function definition, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Definition Return field is undefined",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.definition.ValidateImplementation(context.Background())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
