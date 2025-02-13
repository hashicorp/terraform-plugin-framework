// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDefinitionValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		definition function.Definition
		expected   function.DefinitionValidateResponse
	}{
		"valid-no-params": {
			definition: function.Definition{
				Return: function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{},
		},
		"missing-variadic-param-name": {
			definition: function.Definition{
				VariadicParameter: function.StringParameter{},
				Return:            function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"test-function\" - The variadic parameter does not have a name",
					),
				},
			},
		},
		"missing-param-names": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{},
					function.StringParameter{},
				},
				Return: function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"test-function\" - Parameter at position 0 does not have a name",
					),
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"test-function\" - Parameter at position 1 does not have a name",
					),
				},
			},
		},
		"missing-param-names-with-variadic": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{},
				},
				VariadicParameter: function.NumberParameter{},
				Return:            function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"test-function\" - Parameter at position 0 does not have a name",
					),
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"test-function\" - The variadic parameter does not have a name",
					),
				},
			},
		},
		"result-missing": {
			definition: function.Definition{},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"test-function\" - Definition Return field is undefined",
					),
				},
			},
		},
		"param-dynamic-in-collection": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.MapParameter{
						Name:        "map_with_dynamic",
						ElementType: types.DynamicType,
					},
				},
				Return: function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter \"map_with_dynamic\" at position 0 contains a collection type with a nested dynamic type.\n\n"+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework. "+
							"If underlying dynamic values are required, replace the \"map_with_dynamic\" parameter definition with DynamicParameter instead.",
					),
				},
			},
		},
		"variadic-param-dynamic-in-collection": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{
						Name: "string_param1",
					},
					function.StringParameter{
						Name: "string_param2",
					},
				},
				VariadicParameter: function.SetParameter{
					Name:        "set_with_dynamic",
					ElementType: types.DynamicType,
				},
				Return: function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Variadic parameter \"set_with_dynamic\" contains a collection type with a nested dynamic type.\n\n"+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework. "+
							"If underlying dynamic values are required, replace the variadic parameter definition with DynamicParameter instead.",
					),
				},
			},
		},
		"return-dynamic-in-collection": {
			definition: function.Definition{
				Return: function.ListReturn{
					ElementType: types.DynamicType,
				},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Return contains a collection type with a nested dynamic type.\n\n"+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework. "+
							"If underlying dynamic values are required, replace the return definition with DynamicReturn instead.",
					),
				},
			},
		},
		"conflicting-param-names": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{
						Name: "string-param",
					},
					function.Float64Parameter{
						Name: "float-param",
					},
					function.Int64Parameter{
						Name: "param-dup",
					},
					function.NumberParameter{
						Name: "number-param",
					},
					function.BoolParameter{
						Name: "param-dup",
					},
				},
				Return: function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter names must be unique. "+
							"Function \"test-function\" - Parameters at position 2 and 4 have the same name \"param-dup\"",
					),
				},
			},
		},
		"conflicting-param-names-variadic": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.Float64Parameter{
						Name: "float-param",
					},
					function.Int64Parameter{
						Name: "param-dup",
					},
					function.NumberParameter{
						Name: "number-param",
					},
				},
				VariadicParameter: function.BoolParameter{
					Name: "param-dup",
				},
				Return: function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter names must be unique. "+
							"Function \"test-function\" - Parameter at position 1 and the variadic parameter have the same name \"param-dup\"",
					),
				},
			},
		},
		"conflicting-param-names-variadic-multiple": {
			definition: function.Definition{
				Parameters: []function.Parameter{
					function.StringParameter{
						Name: "param-dup",
					},
					function.Float64Parameter{
						Name: "float-param",
					},
					function.Int64Parameter{
						Name: "param-dup",
					},
					function.NumberParameter{
						Name: "number-param",
					},
					function.BoolParameter{
						Name: "param-dup",
					},
				},
				VariadicParameter: function.BoolParameter{
					Name: "param-dup",
				},
				Return: function.StringReturn{},
			},
			expected: function.DefinitionValidateResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter names must be unique. "+
							"Function \"test-function\" - Parameters at position 0 and 2 have the same name \"param-dup\"",
					),
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter names must be unique. "+
							"Function \"test-function\" - Parameters at position 0 and 4 have the same name \"param-dup\"",
					),
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter names must be unique. "+
							"Function \"test-function\" - Parameter at position 0 and the variadic parameter have the same name \"param-dup\"",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.DefinitionValidateResponse{}

			testCase.definition.ValidateImplementation(context.Background(), function.DefinitionValidateRequest{FuncName: "test-function"}, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
