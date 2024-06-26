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

func TestInt32ReturnGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int32Return
		expected  attr.Type
	}{
		"unset": {
			parameter: function.Int32Return{},
			expected:  basetypes.Int32Type{},
		},
		"CustomType": {
			parameter: function.Int32Return{
				CustomType: testtypes.Int32TypeWithSemanticEquals{},
			},
			expected: testtypes.Int32TypeWithSemanticEquals{},
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
