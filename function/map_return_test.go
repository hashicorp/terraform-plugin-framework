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

func TestMapReturnGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.MapReturn
		expected  attr.Type
	}{
		"ElementType": {
			parameter: function.MapReturn{
				ElementType: basetypes.StringType{},
			},
			expected: basetypes.MapType{
				ElemType: basetypes.StringType{},
			},
		},
		"CustomType": {
			parameter: function.MapReturn{
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

func TestMapReturnValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		returnDef function.MapReturn
		request   function.ValidateReturnImplementationRequest
		expected  *function.ValidateReturnImplementationResponse
	}{
		"customtype": {
			returnDef: function.MapReturn{
				CustomType: testtypes.MapType{},
			},
			request:  function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{},
		},
		"elementtype": {
			returnDef: function.MapReturn{
				ElementType: types.StringType,
			},
			request:  function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{},
		},
		"elementtype-dynamic": {
			returnDef: function.MapReturn{
				ElementType: types.DynamicType,
			},
			request: function.ValidateReturnImplementationRequest{},
			expected: &function.ValidateReturnImplementationResponse{
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
