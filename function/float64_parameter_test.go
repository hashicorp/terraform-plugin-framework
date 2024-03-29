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

func TestFloat64ParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Float64Parameter
		expected  bool
	}{
		"unset": {
			parameter: function.Float64Parameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.Float64Parameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.Float64Parameter{
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

func TestFloat64ParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Float64Parameter
		expected  bool
	}{
		"unset": {
			parameter: function.Float64Parameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.Float64Parameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.Float64Parameter{
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

func TestFloat64ParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Float64Parameter
		expected  string
	}{
		"unset": {
			parameter: function.Float64Parameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.Float64Parameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.Float64Parameter{
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

func TestFloat64ParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Float64Parameter
		expected  string
	}{
		"unset": {
			parameter: function.Float64Parameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.Float64Parameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.Float64Parameter{
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

func TestFloat64ParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Float64Parameter
		expected  string
	}{
		"unset": {
			parameter: function.Float64Parameter{},
			expected:  "",
		},
		"Name-nonempty": {
			parameter: function.Float64Parameter{
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

func TestFloat64ParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Float64Parameter
		expected  attr.Type
	}{
		"unset": {
			parameter: function.Float64Parameter{},
			expected:  basetypes.Float64Type{},
		},
		"CustomType": {
			parameter: function.Float64Parameter{
				CustomType: testtypes.Float64TypeWithSemanticEquals{},
			},
			expected: testtypes.Float64TypeWithSemanticEquals{},
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

func TestFloat64ParameterFloat64Validators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Float64Parameter
		expected  []function.Float64Validator
	}{
		"unset": {
			parameter: function.Float64Parameter{},
			expected:  nil,
		},
		"Validators - empty": {
			parameter: function.Float64Parameter{
				Validators: []function.Float64Validator{}},
			expected: []function.Float64Validator{},
		},
		"Validators": {
			parameter: function.Float64Parameter{
				Validators: []function.Float64Validator{
					testvalidator.Float64{},
				}},
			expected: []function.Float64Validator{
				testvalidator.Float64{},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.Float64Validators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
