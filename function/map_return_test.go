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
