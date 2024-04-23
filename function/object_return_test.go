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
		request   fwfunction.ValidateReturnImplementationRequest
		expected  *fwfunction.ValidateReturnImplementationResponse
	}{
		"customtype": {
			returnDef: function.ObjectReturn{
				CustomType: testtypes.ObjectType{},
			},
			request:  fwfunction.ValidateReturnImplementationRequest{},
			expected: &fwfunction.ValidateReturnImplementationResponse{},
		},
		"attributetypes": {
			returnDef: function.ObjectReturn{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
			},
			request:  fwfunction.ValidateReturnImplementationRequest{},
			expected: &fwfunction.ValidateReturnImplementationResponse{},
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
			request:  fwfunction.ValidateReturnImplementationRequest{},
			expected: &fwfunction.ValidateReturnImplementationResponse{},
		},
		"attributetypes-nested-collection-dynamic": {
			returnDef: function.ObjectReturn{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.ListType{
						ElemType: types.DynamicType,
					},
				},
			},
			request: fwfunction.ValidateReturnImplementationRequest{},
			expected: &fwfunction.ValidateReturnImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Return contains a collection type with a nested dynamic type.\n\n"+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework. "+
							"If underlying dynamic values are required, replace the return definition with DynamicReturn instead.",
					),
				},
			},
		},
		"attributetypes-missing": {
			returnDef: function.ObjectReturn{
				// AttributeTypes intentionally missing
			},
			request:  fwfunction.ValidateReturnImplementationRequest{},
			expected: &fwfunction.ValidateReturnImplementationResponse{
				// No diagnostics are expected as objects can be empty
			},
		},
		"attributetypes-missing-underlying-type": {
			returnDef: function.ObjectReturn{
				AttributeTypes: map[string]attr.Type{
					"nil": nil,
				},
			},
			request: fwfunction.ValidateReturnImplementationRequest{},
			expected: &fwfunction.ValidateReturnImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Function Definition",
						"When validating the function definition, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"Return is missing underlying type.\n\n"+
							"Collection element and object attribute types are always required in Terraform.",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwfunction.ValidateReturnImplementationResponse{}
			testCase.returnDef.ValidateImplementation(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
