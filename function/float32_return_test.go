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

func TestFloat32ReturnGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		parameter function.Float32Return
		expected  attr.Type
	}{
		"unset": {
			parameter: function.Float32Return{},
			expected:  basetypes.Float32Type{},
		},
		"CustomType": {
			parameter: function.Float32Return{
				CustomType: testtypes.Float32TypeWithSemanticEquals{},
			},
			expected: testtypes.Float32TypeWithSemanticEquals{},
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
