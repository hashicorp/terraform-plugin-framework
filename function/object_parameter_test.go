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

func TestObjectParameterGetAllowNullValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ObjectParameter
		expected  bool
	}{
		"unset": {
			parameter: function.ObjectParameter{},
			expected:  false,
		},
		"AllowNullValue-false": {
			parameter: function.ObjectParameter{
				AllowNullValue: false,
			},
			expected: false,
		},
		"AllowNullValue-true": {
			parameter: function.ObjectParameter{
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

func TestObjectParameterGetAllowUnknownValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ObjectParameter
		expected  bool
	}{
		"unset": {
			parameter: function.ObjectParameter{},
			expected:  false,
		},
		"AllowUnknownValues-false": {
			parameter: function.ObjectParameter{
				AllowUnknownValues: false,
			},
			expected: false,
		},
		"AllowUnknownValues-true": {
			parameter: function.ObjectParameter{
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

func TestObjectParameterGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ObjectParameter
		expected  string
	}{
		"unset": {
			parameter: function.ObjectParameter{},
			expected:  "",
		},
		"Description-empty": {
			parameter: function.ObjectParameter{
				Description: "",
			},
			expected: "",
		},
		"Description-nonempty": {
			parameter: function.ObjectParameter{
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

func TestObjectParameterGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ObjectParameter
		expected  string
	}{
		"unset": {
			parameter: function.ObjectParameter{},
			expected:  "",
		},
		"MarkdownDescription-empty": {
			parameter: function.ObjectParameter{
				MarkdownDescription: "",
			},
			expected: "",
		},
		"MarkdownDescription-nonempty": {
			parameter: function.ObjectParameter{
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

func TestObjectParameterGetName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ObjectParameter
		expected  string
	}{
		"unset": {
			parameter: function.ObjectParameter{},
			expected:  function.DefaultParameterName,
		},
		"Name-empty": {
			parameter: function.ObjectParameter{
				Name: "",
			},
			expected: function.DefaultParameterName,
		},
		"Name-nonempty": {
			parameter: function.ObjectParameter{
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

func TestObjectParameterGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ObjectParameter
		expected  attr.Type
	}{
		"ElementType": {
			parameter: function.ObjectParameter{
				AttributeTypes: map[string]attr.Type{
					"test": basetypes.StringType{},
				},
			},
			expected: basetypes.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test": basetypes.StringType{},
				},
			},
		},
		// "CustomType": {
		// 	parameter: function.ObjectParameter{
		// 		CustomType: testtypes.ObjectType{},
		// 	},
		// 	expected: testtypes.ObjectType{},
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
