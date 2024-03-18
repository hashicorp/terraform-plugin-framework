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

func TestMapParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.MapParameter
		expected  bool
	}{
		"unset": {
			parameter: function.MapParameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.MapParameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.MapParameter{
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

func TestMapParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.MapParameter
		expected  bool
	}{
		"unset": {
			parameter: function.MapParameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.MapParameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.MapParameter{
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

func TestMapParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.MapParameter
		expected  string
	}{
		"unset": {
			parameter: function.MapParameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.MapParameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.MapParameter{
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

func TestMapParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.MapParameter
		expected  string
	}{
		"unset": {
			parameter: function.MapParameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.MapParameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.MapParameter{
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

func TestMapParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.MapParameter
		expected  string
	}{
		"unset": {
			parameter: function.MapParameter{},
			expected:  "",
		},
		"Name-nonempty": {
			parameter: function.MapParameter{
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

func TestMapParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.MapParameter
		expected  attr.Type
	}{
		"ElementType": {
			parameter: function.MapParameter{
				ElementType: basetypes.StringType{},
			},
			expected: basetypes.MapType{
				ElemType: basetypes.StringType{},
			},
		},
		"CustomType": {
			parameter: function.MapParameter{
				CustomType: testtypes.MapType{
					MapType: basetypes.MapType{
						ElemType: basetypes.StringType{},
					},
				},
			},
			expected: testtypes.MapType{
				MapType: basetypes.MapType{
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

func TestMapParameterValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		param    function.MapParameter
		request  function.ValidateParameterImplementationRequest
		expected *function.ValidateParameterImplementationResponse
	}{
		"customtype": {
			param: function.MapParameter{
				CustomType: testtypes.MapType{},
			},
			request: function.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"elementtype": {
			param: function.MapParameter{
				ElementType: types.StringType,
			},
			request: function.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"elementtype-dynamic": {
			param: function.MapParameter{
				Name:        "testparam",
				ElementType: types.DynamicType,
			},
			request: function.ValidateParameterImplementationRequest{
				Name:              "testparam",
				ParameterPosition: pointer(int64(0)),
			},
			expected: &function.ValidateParameterImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter \"testparam\" at position 0 contains a collection type with a nested dynamic type.\n\n"+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework. "+
							"If underlying dynamic values are required, replace the \"testparam\" parameter definition with DynamicParameter instead.",
					),
				},
			},
		},
		"elementtype-dynamic-variadic": {
			param: function.MapParameter{
				Name:        "testparam",
				ElementType: types.DynamicType,
			},
			request: function.ValidateParameterImplementationRequest{
				Name: "testparam",
			},
			expected: &function.ValidateParameterImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Variadic parameter \"testparam\" contains a collection type with a nested dynamic type.\n\n"+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework. "+
							"If underlying dynamic values are required, replace the variadic parameter definition with DynamicParameter instead.",
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
