// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestGetFunctionsResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.GetFunctionsResponse
		expected *tfprotov6.GetFunctionsResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"diagnostics": {
			input: &fwserver.GetFunctionsResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("warning summary", "warning detail"),
					diag.NewErrorDiagnostic("error summary", "error detail"),
				},
			},
			expected: &tfprotov6.GetFunctionsResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
				Functions: map[string]*tfprotov6.Function{},
			},
		},
		"functions": {
			input: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction1": {
						Return: function.StringReturn{},
					},
					"testfunction2": {
						Return: function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetFunctionsResponse{
				Functions: map[string]*tfprotov6.Function{
					"testfunction1": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
					"testfunction2": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
			},
		},
		"functions-deprecationmessage": {
			input: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						DeprecationMessage: "test deprecation message",
						Return:             function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetFunctionsResponse{
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						DeprecationMessage: "test deprecation message",
						Parameters:         []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
			},
		},
		"functions-description": {
			input: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Description: "test description",
						Return:      function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetFunctionsResponse{
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Description: "test description",
						Parameters:  []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
			},
		},
		"functions-parameters": {
			input: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Parameters: []function.Parameter{
							function.BoolParameter{},
							function.Int64Parameter{},
							function.StringParameter{},
						},
						Return: function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetFunctionsResponse{
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Parameters: []*tfprotov6.FunctionParameter{
							{
								Name: function.DefaultParameterName,
								Type: tftypes.Bool,
							},
							{
								Name: function.DefaultParameterName,
								Type: tftypes.Number,
							},
							{
								Name: function.DefaultParameterName,
								Type: tftypes.String,
							},
						},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
			},
		},
		"functions-result": {
			input: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Return: function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetFunctionsResponse{
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
			},
		},
		"functions-summary": {
			input: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Return:  function.StringReturn{},
						Summary: "test summary",
					},
				},
			},
			expected: &tfprotov6.GetFunctionsResponse{
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
						Summary: "test summary",
					},
				},
			},
		},
		"functions-variadicparameter": {
			input: &fwserver.GetFunctionsResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Return:            function.StringReturn{},
						VariadicParameter: function.StringParameter{},
					},
				},
			},
			expected: &tfprotov6.GetFunctionsResponse{
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
						VariadicParameter: &tfprotov6.FunctionParameter{
							Name: function.DefaultParameterName,
							Type: tftypes.String,
						},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.GetFunctionsResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
