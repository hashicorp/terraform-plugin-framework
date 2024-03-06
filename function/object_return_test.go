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

func TestObjectReturnGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.ObjectReturn
		expected  attr.Type
	}{
		"ElementType": {
			parameter: function.ObjectReturn{
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
			parameter: function.ObjectReturn{
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

func TestObjectReturnValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		returnDef function.ObjectReturn
		request   function.ValidateReturnImplementationRequest
		expected  *function.ValidateReturnImplementationResponse
	}{
		"customtype": {
			returnDef: function.ObjectReturn{
				CustomType: testtypes.ObjectType{},
			},
			request:  function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{},
		},
		"attributetypes": {
			returnDef: function.ObjectReturn{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
			},
			request:  function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{},
		},
		"attributetypes-dynamic": {
			returnDef: function.ObjectReturn{
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
			request:  function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{},
		},
		"attributetypes-nested-collection-dynamic": {
			returnDef: function.ObjectReturn{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.ListType{
						ElemType: types.DynamicType,
					},
				},
			},
			request: function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Return contains a collection type with a nested dynamic type. "+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &function.ValidateReturnImplementationResponse{}
			testCase.returnDef.ValidateImplementation(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
