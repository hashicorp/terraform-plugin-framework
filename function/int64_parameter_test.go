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
	"github.com/hashicorp/terraform-plugin-framework/internal/fwfunction"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestInt64ParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int64Parameter
		expected  bool
	}{
		"unset": {
			parameter: function.Int64Parameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.Int64Parameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.Int64Parameter{
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

func TestInt64ParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int64Parameter
		expected  bool
	}{
		"unset": {
			parameter: function.Int64Parameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.Int64Parameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.Int64Parameter{
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

func TestInt64ParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int64Parameter
		expected  string
	}{
		"unset": {
			parameter: function.Int64Parameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.Int64Parameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.Int64Parameter{
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

func TestInt64ParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int64Parameter
		expected  string
	}{
		"unset": {
			parameter: function.Int64Parameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.Int64Parameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.Int64Parameter{
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

func TestInt64ParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int64Parameter
		expected  string
	}{
		"unset": {
			parameter: function.Int64Parameter{},
			expected:  "",
		},
		"Name-nonempty": {
			parameter: function.Int64Parameter{
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

func TestInt64ParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int64Parameter
		expected  attr.Type
	}{
		"unset": {
			parameter: function.Int64Parameter{},
			expected:  basetypes.Int64Type{},
		},
		"CustomType": {
			parameter: function.Int64Parameter{
				CustomType: testtypes.Int64TypeWithSemanticEquals{},
			},
			expected: testtypes.Int64TypeWithSemanticEquals{},
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

func TestInt64ParameterInt64Validators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int64Parameter
		expected  []function.Int64ParameterValidator
	}{
		"unset": {
			parameter: function.Int64Parameter{},
			expected:  nil,
		},
		"Validators - empty": {
			parameter: function.Int64Parameter{
				Validators: []function.Int64ParameterValidator{}},
			expected: []function.Int64ParameterValidator{},
		},
		"Validators": {
			parameter: function.Int64Parameter{
				Validators: []function.Int64ParameterValidator{
					testvalidator.Int64{},
				}},
			expected: []function.Int64ParameterValidator{
				testvalidator.Int64{},
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

func TestInt64ParameterValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		param    function.Int64Parameter
		request  fwfunction.ValidateParameterImplementationRequest
		expected *fwfunction.ValidateParameterImplementationResponse
	}{
		"name": {
			param: function.Int64Parameter{
				Name: "testparam",
			},
			request: fwfunction.ValidateParameterImplementationRequest{
				FunctionName:      "testfunc",
				ParameterPosition: pointer(int64(0)),
			},
			expected: &fwfunction.ValidateParameterImplementationResponse{},
		},
		"name-missing": {
			param: function.Int64Parameter{
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
