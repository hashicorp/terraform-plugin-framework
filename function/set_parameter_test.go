// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestSetParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.SetParameter
		expected  bool
	}{
		"unset": {
			parameter: function.SetParameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.SetParameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.SetParameter{
				AllowNullValue: true,
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetAllowNullValue()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.SetParameter
		expected  bool
	}{
		"unset": {
			parameter: function.SetParameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.SetParameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.SetParameter{
				AllowUnknownValues: true,
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetAllowUnknownValues()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.SetParameter
		expected  string
	}{
		"unset": {
			parameter: function.SetParameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.SetParameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.SetParameter{
				Description: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.SetParameter
		expected  string
	}{
		"unset": {
			parameter: function.SetParameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.SetParameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.SetParameter{
				MarkdownDescription: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetMarkdownDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.SetParameter
		expected  string
	}{
		"unset": {
			parameter: function.SetParameter{},
			expected:  "",
		},
		"Name-nonempty": {
			parameter: function.SetParameter{
				Name: "test",
			},
			expected: "test",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetName()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.SetParameter
		expected  attr.Type
	}{
		"ElementType": {
			parameter: function.SetParameter{
				ElementType: basetypes.StringType{},
			},
			expected: basetypes.SetType{
				ElemType: basetypes.StringType{},
			},
		},
		"CustomType": {
			parameter: function.SetParameter{
				CustomType: testtypes.SetType{
					SetType: basetypes.SetType{
						ElemType: basetypes.StringType{},
					},
				},
			},
			expected: testtypes.SetType{
				SetType: basetypes.SetType{
					ElemType: basetypes.StringType{},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetType()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetParameterValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		param    function.SetParameter
		request  function.ValidateParameterImplementationRequest
		expected *function.ValidateParameterImplementationResponse
	}{
		"customtype": {
			param: function.SetParameter{
				CustomType: testtypes.SetType{},
			},
			request: function.ValidateParameterImplementationRequest{
				FunctionArgument: 0,
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"elementtype": {
			param: function.SetParameter{
				ElementType: types.StringType,
			},
			request: function.ValidateParameterImplementationRequest{
				FunctionArgument: 0,
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"elementtype-dynamic": {
			param: function.SetParameter{
				Name:        "testparam",
				ElementType: types.DynamicType,
			},
			request: function.ValidateParameterImplementationRequest{
				Name:             "testparam",
				FunctionArgument: 0,
			},
			expected: &function.ValidateParameterImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter \"testparam\" at position 0 contains a collection type with a nested dynamic type. "+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &function.ValidateParameterImplementationResponse{}
			testCase.param.ValidateImplementation(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
