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
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			expected:  "",
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
		"CustomType": {
			parameter: function.ObjectParameter{
				CustomType: testtypes.ObjectType{
					ObjectType: basetypes.ObjectType{
						AttrTypes: map[string]attr.Type{
							"test": basetypes.StringType{},
						},
					},
				},
			},
			expected: testtypes.ObjectType{
				ObjectType: basetypes.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test": basetypes.StringType{},
					},
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

func TestObjectParameterValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		param    function.ObjectParameter
		request  function.ValidateParameterImplementationRequest
		expected *function.ValidateParameterImplementationResponse
	}{
		"customtype": {
			param: function.ObjectParameter{
				CustomType: testtypes.ObjectType{},
			},
			request: function.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"attributetypes": {
			param: function.ObjectParameter{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
			},
			request: function.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"attributetypes-dynamic": {
			param: function.ObjectParameter{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.DynamicType,
					"test_list": types.ListType{
						ElemType: types.StringType,
					},
					"test_obj": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr": types.DynamicType,
						},
					},
				},
			},
			request: function.ValidateParameterImplementationRequest{
				ParameterPosition: pointer(int64(0)),
			},
			expected: &function.ValidateParameterImplementationResponse{},
		},
		"attributetypes-nested-collection-dynamic": {
			param: function.ObjectParameter{
				Name: "testparam",
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.ListType{
						ElemType: types.DynamicType,
					},
				},
			},
			request: function.ValidateParameterImplementationRequest{
				Name:              "testparam",
				ParameterPosition: pointer(int64(0)),
			},
			expected: &function.ValidateParameterImplementationResponse{
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
		"attributetypes-nested-collection-dynamic-variadic": {
			param: function.ObjectParameter{
				Name: "testparam",
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.ListType{
						ElemType: types.DynamicType,
					},
				},
			},
			request: function.ValidateParameterImplementationRequest{
				Name: "testparam",
			},
			expected: &function.ValidateParameterImplementationResponse{
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
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &function.ValidateParameterImplementationResponse{}
			testCase.param.ValidateImplementation(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
