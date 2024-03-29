// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
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

func TestDynamicParameterDynamicValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.DynamicParameter
		expected  []function.DynamicValidator
	}{
		"unset": {
			parameter: function.DynamicParameter{},
			expected:  nil,
		},
		"Validators - empty": {
			parameter: function.DynamicParameter{
				Validators: []function.DynamicValidator{}},
			expected: []function.DynamicValidator{},
		},
		"Validators": {
			parameter: function.DynamicParameter{
				Validators: []function.DynamicValidator{
					testvalidator.Dynamic{},
				}},
			expected: []function.DynamicValidator{
				testvalidator.Dynamic{},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.DynamicValidators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
