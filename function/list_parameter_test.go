// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
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
			expected:  function.DefaultParameterName,
		},
		"Name-empty": {
			parameter: function.ListParameter{
				Name: "",
			},
			expected: function.DefaultParameterName,
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
		// "CustomType": {
		// 	parameter: function.ListParameter{
		// 		CustomType: testtypes.ListType{},
		// 	},
		// 	expected: testtypes.ListType{},
		// },
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
