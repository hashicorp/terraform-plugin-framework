// Copyright IBM Corp. 2021, 2026
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

func TestInt64ReturnGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Int64Return
		expected  attr.Type
	}{
		"unset": {
			parameter: function.Int64Return{},
			expected:  basetypes.Int64Type{},
		},
		"CustomType": {
			parameter: function.Int64Return{
				CustomType: testtypes.Int64TypeWithSemanticEquals{},
			},
			expected: testtypes.Int64TypeWithSemanticEquals{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.parameter.GetType()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
