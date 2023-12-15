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
			expected:  function.DefaultParameterName,
		},
		"Name-empty": {
			parameter: function.MapParameter{
				Name: "",
			},
			expected: function.DefaultParameterName,
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
