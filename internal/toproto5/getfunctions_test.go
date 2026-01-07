// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
)

func TestGetFunctionsResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.GetFunctionsResponse
		expected *tfprotov5.GetFunctionsResponse
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
			expected: &tfprotov5.GetFunctionsResponse{
				Diagnostics: []*tfprotov5.Diagnostic{
					{
						Severity: tfprotov5.DiagnosticSeverityWarning,
						Summary:  "warning summary",
						Detail:   "warning detail",
					},
					{
						Severity: tfprotov5.DiagnosticSeverityError,
						Summary:  "error summary",
						Detail:   "error detail",
					},
				},
				Functions: map[string]*tfprotov5.Function{},
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
			expected: &tfprotov5.GetFunctionsResponse{
				Functions: map[string]*tfprotov5.Function{
					"testfunction1": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
							Type: tftypes.String,
						},
					},
					"testfunction2": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
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
			expected: &tfprotov5.GetFunctionsResponse{
				Functions: map[string]*tfprotov5.Function{
					"testfunction": {
						DeprecationMessage: "test deprecation message",
						Parameters:         []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
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
			expected: &tfprotov5.GetFunctionsResponse{
				Functions: map[string]*tfprotov5.Function{
					"testfunction": {
						Description: "test description",
						Parameters:  []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
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
			expected: &tfprotov5.GetFunctionsResponse{
				Functions: map[string]*tfprotov5.Function{
					"testfunction": {
						Parameters: []*tfprotov5.FunctionParameter{
							{
								Type: tftypes.Bool,
							},
							{
								Type: tftypes.Number,
							},
							{
								Type: tftypes.String,
							},
						},
						Return: &tfprotov5.FunctionReturn{
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
			expected: &tfprotov5.GetFunctionsResponse{
				Functions: map[string]*tfprotov5.Function{
					"testfunction": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
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
			expected: &tfprotov5.GetFunctionsResponse{
				Functions: map[string]*tfprotov5.Function{
					"testfunction": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
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
			expected: &tfprotov5.GetFunctionsResponse{
				Functions: map[string]*tfprotov5.Function{
					"testfunction": {
						Parameters: []*tfprotov5.FunctionParameter{},
						Return: &tfprotov5.FunctionReturn{
							Type: tftypes.String,
						},
						VariadicParameter: &tfprotov5.FunctionParameter{
							Type: tftypes.String,
						},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.GetFunctionsResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
