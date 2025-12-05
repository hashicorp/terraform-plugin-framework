// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwfunction"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetType()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetParameterSetValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.SetParameter
		expected  []function.SetParameterValidator
	}{
		"unset": {
			parameter: function.SetParameter{},
			expected:  nil,
		},
		"Validators - empty": {
			parameter: function.SetParameter{
				Validators: []function.SetParameterValidator{}},
			expected: []function.SetParameterValidator{},
		},
		"Validators": {
			parameter: function.SetParameter{
				Validators: []function.SetParameterValidator{
					testvalidator.Set{},
				}},
			expected: []function.SetParameterValidator{
				testvalidator.Set{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetValidators()

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
		request  fwfunction.ValidateParameterImplementationRequest
		expected *fwfunction.ValidateParameterImplementationResponse
	}{
		"customtype": {
			param: function.SetParameter{
				Name:       "testparam",
				CustomType: testtypes.SetType{},
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{},
		},
		"elementtype": {
			param: function.SetParameter{
				Name:        "testparam",
				ElementType: types.StringType,
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{},
		},
		"elementtype-dynamic": {
			param: function.SetParameter{
				Name:        "testparam",
				ElementType: types.DynamicType,
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{
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
			param: function.SetParameter{
				Name:        "testparam",
				ElementType: types.DynamicType,
			},
			request: fwfunction.ValidateParameterImplementationRequest{},
			expected: &fwfunction.ValidateParameterImplementationResponse{
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
		"elementtype-missing": {
			param: function.SetParameter{
				Name: "testparam",
				// ElementType intentionally missing
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Parameter \"testparam\" at position 0 is missing underlying type.\n\n"+
							"Collection element and object attribute types are always required in Terraform.",
					),
				},
			},
		},
		"name": {
			param: function.SetParameter{
				Name:        "testparam",
				ElementType: types.StringType,
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				FunctionName:      "testfunc",
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{},
		},
		"name-missing": {
			param: function.SetParameter{
				// Name intentionally missing
				ElementType: types.StringType,
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				FunctionName:      "testfunc",
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Function \"testfunc\" - Parameter at position 0 does not have a name",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwfunction.ValidateParameterImplementationResponse{}
			testCase.param.ValidateImplementation(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
