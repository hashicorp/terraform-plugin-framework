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

func TestListParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ListParameter
		expected  bool
	}{
		"unset": {
			parameter: function.ListParameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.ListParameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.ListParameter{
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

func TestListParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ListParameter
		expected  bool
	}{
		"unset": {
			parameter: function.ListParameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.ListParameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.ListParameter{
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

func TestListParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ListParameter
		expected  string
	}{
		"unset": {
			parameter: function.ListParameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.ListParameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.ListParameter{
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

func TestListParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ListParameter
		expected  string
	}{
		"unset": {
			parameter: function.ListParameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.ListParameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.ListParameter{
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

func TestListParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ListParameter
		expected  string
	}{
		"unset": {
			parameter: function.ListParameter{},
			expected:  "",
		},
		"Name-nonempty": {
			parameter: function.ListParameter{
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

func TestListParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ListParameter
		expected  attr.Type
	}{
		"ElementType": {
			parameter: function.ListParameter{
				ElementType: basetypes.StringType{},
			},
			expected: basetypes.ListType{
				ElemType: basetypes.StringType{},
			},
		},
		"CustomType": {
			parameter: function.ListParameter{
				CustomType: testtypes.ListType{
					ListType: basetypes.ListType{
						ElemType: basetypes.StringType{},
					},
				},
			},
			expected: testtypes.ListType{
				ListType: basetypes.ListType{
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

func TestListParameterValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		param    function.ListParameter
		request  function.ValidateParameterImplementationRequest
		expected *function.ValidateParameterImplementationResponse
	}{
		"customtype": {
			param: function.ListParameter{
				CustomType: testtypes.ListType{},
			},
			request: function.ValidateParameterImplementationRequest{
				FunctionArgument: 0,
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"elementtype": {
			param: function.ListParameter{
				ElementType: types.StringType,
			},
			request: function.ValidateParameterImplementationRequest{
				FunctionArgument: 0,
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"elementtype-dynamic": {
			param: function.ListParameter{
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
