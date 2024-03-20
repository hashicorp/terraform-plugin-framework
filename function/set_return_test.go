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

func TestSetReturnGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.SetReturn
		expected  attr.Type
	}{
		"ElementType": {
			parameter: function.SetReturn{
				ElementType: basetypes.StringType{},
			},
			expected: basetypes.SetType{
				ElemType: basetypes.StringType{},
			},
		},
		"CustomType": {
			parameter: function.SetReturn{
				CustomType: testtypes.SetType{
					SetType: basetypes.SetType{
						ElemType: basetypes.StringType{},
					},
				},
			},
			expected: testtypes.SetType{
				SetType: basetypes.SetType{
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

func TestSetReturnValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		returnDef function.SetReturn
		request   fwfunction.ValidateReturnImplementationRequest
		expected  *fwfunction.ValidateReturnImplementationResponse
	}{
		"customtype": {
			returnDef: function.SetReturn{
				CustomType: testtypes.SetType{},
			},
			request:  fwfunction.ValidateReturnImplementationRequest{},
			expected: &fwfunction.ValidateReturnImplementationResponse{},
		},
		"elementtype": {
			returnDef: function.SetReturn{
				ElementType: types.StringType,
			},
			request:  fwfunction.ValidateReturnImplementationRequest{},
			expected: &fwfunction.ValidateReturnImplementationResponse{},
		},
		"elementtype-dynamic": {
			returnDef: function.SetReturn{
				ElementType: types.DynamicType,
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
