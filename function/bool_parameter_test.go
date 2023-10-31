// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestBoolParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.BoolParameter
		expected  bool
	}{
		"unset": {
			parameter: function.BoolParameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.BoolParameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.BoolParameter{
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

func TestBoolParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.BoolParameter
		expected  bool
	}{
		"unset": {
			parameter: function.BoolParameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.BoolParameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.BoolParameter{
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

func TestBoolParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.BoolParameter
		expected  string
	}{
		"unset": {
			parameter: function.BoolParameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.BoolParameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.BoolParameter{
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

func TestBoolParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.BoolParameter
		expected  string
	}{
		"unset": {
			parameter: function.BoolParameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.BoolParameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.BoolParameter{
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

func TestBoolParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.BoolParameter
		expected  string
	}{
		"unset": {
			parameter: function.BoolParameter{},
			expected:  function.DefaultParameterName,
		},
		"Name-empty": {
			parameter: function.BoolParameter{
				Name: "",
			},
			expected: function.DefaultParameterName,
		},
		"Name-nonempty": {
			parameter: function.BoolParameter{
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

func TestBoolParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.BoolParameter
		expected  attr.Type
	}{
		"unset": {
			parameter: function.BoolParameter{},
			expected:  basetypes.BoolType{},
		},
		"CustomType": {
			parameter: function.BoolParameter{
				CustomType: testtypes.BoolType{},
			},
			expected: testtypes.BoolType{},
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
