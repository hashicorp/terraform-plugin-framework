// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/fwerror"
)

func TestFunctionError_Equal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErr  fwerror.FunctionError
		other    fwerror.FunctionError
		expected bool
	}{
		"matching": {
			funcErr:  fwerror.NewFunctionError("test summary: test detail"),
			other:    fwerror.NewFunctionError("test summary: test detail"),
			expected: true,
		},
		"nil": {
			funcErr:  fwerror.NewFunctionError("test summary: test detail"),
			other:    nil,
			expected: false,
		},
		"different-msg": {
			funcErr:  fwerror.NewFunctionError("test summary: test detail"),
			other:    fwerror.NewFunctionError("test summary: different detail"),
			expected: false,
		},
		"different-type": {
			funcErr:  fwerror.NewFunctionError("test summary: test detail"),
			other:    fwerror.NewArgumentFunctionError(0, "test summary: test detail"),
			expected: false,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.funcErr.Equal(tc.other)

			if got != tc.expected {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}
		})
	}
}
