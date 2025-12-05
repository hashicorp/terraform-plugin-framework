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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestStringParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.StringParameter
		expected  bool
	}{
		"unset": {
			parameter: function.StringParameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.StringParameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.StringParameter{
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

func TestStringParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.StringParameter
		expected  bool
	}{
		"unset": {
			parameter: function.StringParameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.StringParameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.StringParameter{
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

func TestStringParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.StringParameter
		expected  string
	}{
		"unset": {
			parameter: function.StringParameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.StringParameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.StringParameter{
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

func TestStringParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.StringParameter
		expected  string
	}{
		"unset": {
			parameter: function.StringParameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.StringParameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.StringParameter{
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

func TestStringParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.StringParameter
		expected  string
	}{
		"unset": {
			parameter: function.StringParameter{},
			expected:  "",
		},
		"Name-nonempty": {
			parameter: function.StringParameter{
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

func TestStringParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.StringParameter
		expected  attr.Type
	}{
		"unset": {
			parameter: function.StringParameter{},
			expected:  basetypes.StringType{},
		},
		"CustomType": {
			parameter: function.StringParameter{
				CustomType: testtypes.StringType{},
			},
			expected: testtypes.StringType{},
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

func TestStringParameterStringValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.StringParameter
		expected  []function.StringParameterValidator
	}{
		"unset": {
			parameter: function.StringParameter{},
			expected:  nil,
		},
		"Validators - empty": {
			parameter: function.StringParameter{
				Validators: []function.StringParameterValidator{}},
			expected: []function.StringParameterValidator{},
		},
		"Validators": {
			parameter: function.StringParameter{
				Validators: []function.StringParameterValidator{
					testvalidator.String{},
				}},
			expected: []function.StringParameterValidator{
				testvalidator.String{},
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

func TestStringParameterValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		param    function.StringParameter
		request  fwfunction.ValidateParameterImplementationRequest
		expected *fwfunction.ValidateParameterImplementationResponse
	}{
		"name": {
			param: function.StringParameter{
				Name: "testparam",
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				FunctionName:      "testfunc",
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{},
		},
		"name-missing": {
			param: function.StringParameter{
				// Name intentionally missing
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
