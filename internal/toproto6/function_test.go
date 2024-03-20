// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestFunction(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       function.Definition
		expected *tfprotov6.Function
	}{
		"deprecationmessage": {
			fw: function.Definition{
				DeprecationMessage: "test deprecation message",
				Return:             function.StringReturn{},
			},
			expected: &tfprotov6.Function{
				DeprecationMessage: "test deprecation message",
				Parameters:         []*tfprotov6.FunctionParameter{},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
			},
		},
		"description": {
			fw: function.Definition{
				Description: "test description",
				Return:      function.StringReturn{},
			},
			expected: &tfprotov6.Function{
				Description:     "test description",
				DescriptionKind: tfprotov6.StringKindPlain,
				Parameters:      []*tfprotov6.FunctionParameter{},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
			},
		},
		"description-markdown": {
			fw: function.Definition{
				MarkdownDescription: "test description",
				Return:              function.StringReturn{},
			},
			expected: &tfprotov6.Function{
				Description:     "test description",
				DescriptionKind: tfprotov6.StringKindMarkdown,
				Parameters:      []*tfprotov6.FunctionParameter{},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
			},
		},
		"parameters": {
			fw: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						Name: "bool",
					},
					function.Int64Parameter{
						Name: "int64",
					},
					function.StringParameter{
						Name: "string",
					},
				},
				Return: function.StringReturn{},
			},
			expected: &tfprotov6.Function{
				Parameters: []*tfprotov6.FunctionParameter{
					{
						Name: "bool",
						Type: tftypes.Bool,
					},
					{
						Name: "int64",
						Type: tftypes.Number,
					},
					{
						Name: "string",
						Type: tftypes.String,
					},
				},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
			},
		},
		"parameters-with-variadic": {
			fw: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{
						Name: "bool",
					},
					function.Int64Parameter{
						Name: "int64",
					},
					function.StringParameter{
						Name: "string",
					},
				},
				VariadicParameter: function.Float64Parameter{
					Name: "variadic_float64",
				},
				Return: function.StringReturn{},
			},
			expected: &tfprotov6.Function{
				Parameters: []*tfprotov6.FunctionParameter{
					{
						Name: "bool",
						Type: tftypes.Bool,
					},
					{
						Name: "int64",
						Type: tftypes.Number,
					},
					{
						Name: "string",
						Type: tftypes.String,
					},
				},
				VariadicParameter: &tfprotov6.FunctionParameter{
					Name: "variadic_float64",
					Type: tftypes.Number,
				},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
			},
		},
		"parameters-unnamed": {
			fw: function.Definition{
				Parameters: []function.Parameter{
					function.BoolParameter{},
					function.Int64Parameter{},
					function.StringParameter{},
				},
				VariadicParameter: function.Float64Parameter{},
				Return:            function.StringReturn{},
			},
			expected: &tfprotov6.Function{
				Parameters: []*tfprotov6.FunctionParameter{
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
				VariadicParameter: &tfprotov6.FunctionParameter{
					Type: tftypes.Number,
				},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
			},
		},
		"result": {
			fw: function.Definition{
				Return: function.StringReturn{},
			},
			expected: &tfprotov6.Function{
				Parameters: []*tfprotov6.FunctionParameter{},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
			},
		},
		"summary": {
			fw: function.Definition{
				Return:  function.StringReturn{},
				Summary: "test summary",
			},
			expected: &tfprotov6.Function{
				Parameters: []*tfprotov6.FunctionParameter{},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
				Summary: "test summary",
			},
		},
		"variadicparameter": {
			fw: function.Definition{
				Return:            function.StringReturn{},
				VariadicParameter: function.StringParameter{},
			},
			expected: &tfprotov6.Function{
				Parameters: []*tfprotov6.FunctionParameter{},
				Return: &tfprotov6.FunctionReturn{
					Type: tftypes.String,
				},
				VariadicParameter: &tfprotov6.FunctionParameter{
					Type: tftypes.String,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.Function(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFunctionMetadata(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       fwserver.FunctionMetadata
		expected tfprotov6.FunctionMetadata
	}{
		"name": {
			fw: fwserver.FunctionMetadata{
				Name: "test",
			},
			expected: tfprotov6.FunctionMetadata{
				Name: "test",
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.FunctionMetadata(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFunctionParameter(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       function.Parameter
		expected *tfprotov6.FunctionParameter
	}{
		"nil": {
			fw:       nil,
			expected: nil,
		},
		"allownullvalue": {
			fw: function.BoolParameter{
				Name:           "bool",
				AllowNullValue: true,
			},
			expected: &tfprotov6.FunctionParameter{
				AllowNullValue: true,
				Name:           "bool",
				Type:           tftypes.Bool,
			},
		},
		"allowunknownvalues": {
			fw: function.BoolParameter{
				Name:               "bool",
				AllowUnknownValues: true,
			},
			expected: &tfprotov6.FunctionParameter{
				AllowUnknownValues: true,
				Name:               "bool",
				Type:               tftypes.Bool,
			},
		},
		"description": {
			fw: function.BoolParameter{
				Name:        "bool",
				Description: "test description",
			},
			expected: &tfprotov6.FunctionParameter{
				Description:     "test description",
				DescriptionKind: tfprotov6.StringKindPlain,
				Name:            "bool",
				Type:            tftypes.Bool,
			},
		},
		"description-markdown": {
			fw: function.BoolParameter{
				Name:                "bool",
				MarkdownDescription: "test description",
			},
			expected: &tfprotov6.FunctionParameter{
				Description:     "test description",
				DescriptionKind: tfprotov6.StringKindMarkdown,
				Name:            "bool",
				Type:            tftypes.Bool,
			},
		},
		"name": {
			fw: function.BoolParameter{
				Name: "test",
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "test",
				Type: tftypes.Bool,
			},
		},
		"name-empty": {
			fw: function.BoolParameter{},
			expected: &tfprotov6.FunctionParameter{
				Name: "", // default is applied by the toproto6.Function method
				Type: tftypes.Bool,
			},
		},
		"type-bool": {
			fw: function.BoolParameter{
				Name: "bool",
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "bool",
				Type: tftypes.Bool,
			},
		},
		"type-float64": {
			fw: function.Float64Parameter{
				Name: "float64",
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "float64",
				Type: tftypes.Number,
			},
		},
		"type-int64": {
			fw: function.Int64Parameter{
				Name: "int64",
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "int64",
				Type: tftypes.Number,
			},
		},
		"type-list": {
			fw: function.ListParameter{
				Name:        "list",
				ElementType: basetypes.StringType{},
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "list",
				Type: tftypes.List{
					ElementType: tftypes.String,
				},
			},
		},
		"type-map": {
			fw: function.MapParameter{
				Name:        "map",
				ElementType: basetypes.StringType{},
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "map",
				Type: tftypes.Map{
					ElementType: tftypes.String,
				},
			},
		},
		"type-number": {
			fw: function.NumberParameter{
				Name: "number",
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "number",
				Type: tftypes.Number,
			},
		},
		"type-object": {
			fw: function.ObjectParameter{
				Name: "object",
				AttributeTypes: map[string]attr.Type{
					"bool":   basetypes.BoolType{},
					"int64":  basetypes.Int64Type{},
					"string": basetypes.StringType{},
				},
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "object",
				Type: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"bool":   tftypes.Bool,
						"int64":  tftypes.Number,
						"string": tftypes.String,
					},
				},
			},
		},
		"type-set": {
			fw: function.SetParameter{
				Name:        "set",
				ElementType: basetypes.StringType{},
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "set",
				Type: tftypes.Set{
					ElementType: tftypes.String,
				},
			},
		},
		"type-string": {
			fw: function.StringParameter{
				Name: "string",
			},
			expected: &tfprotov6.FunctionParameter{
				Name: "string",
				Type: tftypes.String,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.FunctionParameter(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFunctionReturn(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw       function.Return
		expected *tfprotov6.FunctionReturn
	}{
		"nil": {
			fw:       nil,
			expected: nil,
		},
		"type-string": {
			fw: function.StringReturn{},
			expected: &tfprotov6.FunctionReturn{
				Type: tftypes.String,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.FunctionReturn(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFunctionResultData(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw          function.ResultData
		expected    *tfprotov6.DynamicValue
		expectedErr *function.FuncError
	}{
		"empty": {
			fw:          function.ResultData{},
			expected:    nil,
			expectedErr: nil,
		},
		"value-nil": {
			fw:          function.NewResultData(nil),
			expected:    nil,
			expectedErr: nil,
		},
		"value": {
			fw:          function.NewResultData(basetypes.NewBoolValue(true)),
			expected:    DynamicValueMust(tftypes.NewValue(tftypes.Bool, true)),
			expectedErr: nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := toproto6.FunctionResultData(context.Background(), testCase.fw)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedErr); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
