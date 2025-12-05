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

func TestDynamicParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.DynamicParameter
		expected  bool
	}{
		"unset": {
			parameter: function.DynamicParameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.DynamicParameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.DynamicParameter{
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

func TestDynamicParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.DynamicParameter
		expected  bool
	}{
		"unset": {
			parameter: function.DynamicParameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.DynamicParameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.DynamicParameter{
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

func TestDynamicParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.DynamicParameter
		expected  string
	}{
		"unset": {
			parameter: function.DynamicParameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.DynamicParameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.DynamicParameter{
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

func TestDynamicParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.DynamicParameter
		expected  string
	}{
		"unset": {
			parameter: function.DynamicParameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.DynamicParameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.DynamicParameter{
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

func TestDynamicParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.DynamicParameter
		expected  string
	}{
		"unset": {
			parameter: function.DynamicParameter{},
			expected:  "",
		},
		"Name-nonempty": {
			parameter: function.DynamicParameter{
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

func TestDynamicParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.DynamicParameter
		expected  attr.Type
	}{
		"unset": {
			parameter: function.DynamicParameter{},
			expected:  basetypes.DynamicType{},
		},
		"CustomType": {
			parameter: function.DynamicParameter{
				CustomType: testtypes.DynamicType{},
			},
			expected: testtypes.DynamicType{},
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

func TestDynamicParameterDynamicValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.DynamicParameter
		expected  []function.DynamicParameterValidator
	}{
		"unset": {
			parameter: function.DynamicParameter{},
			expected:  nil,
		},
		"Validators - empty": {
			parameter: function.DynamicParameter{
				Validators: []function.DynamicParameterValidator{}},
			expected: []function.DynamicParameterValidator{},
		},
		"Validators": {
			parameter: function.DynamicParameter{
				Validators: []function.DynamicParameterValidator{
					testvalidator.Dynamic{},
				}},
			expected: []function.DynamicParameterValidator{
				testvalidator.Dynamic{},
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

func TestDynamicParameterValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		param    function.DynamicParameter
		request  fwfunction.ValidateParameterImplementationRequest
		expected *fwfunction.ValidateParameterImplementationResponse
	}{
		"name": {
			param: function.DynamicParameter{
				Name: "testparam",
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				FunctionName:      "testfunc",
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{},
		},
		"name-missing": {
			param: function.DynamicParameter{
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
