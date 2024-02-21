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
