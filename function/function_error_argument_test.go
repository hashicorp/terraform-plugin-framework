// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

func TestArgumentFunctionError_Equal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErr  function.FunctionErrorWithFunctionArgument
		other    function.FunctionError
		expected bool
	}{
		"matching": {
			funcErr:  function.NewArgumentFunctionError(0, "test summary: test detail"),
			other:    function.NewArgumentFunctionError(0, "test summary: test detail"),
			expected: true,
		},
		"nil": {
			funcErr:  function.NewArgumentFunctionError(0, "test summary: test detail"),
			other:    nil,
			expected: false,
		},
		"different-arg": {
			funcErr:  function.NewArgumentFunctionError(0, "test summary: test detail"),
			other:    function.NewArgumentFunctionError(1, "test summary: test detail"),
			expected: false,
		},
		"different-msg": {
			funcErr:  function.NewArgumentFunctionError(0, "test summary: test detail"),
			other:    function.NewArgumentFunctionError(0, "test summary: different detail"),
			expected: false,
		},
		"different-type": {
			funcErr:  function.NewArgumentFunctionError(0, "test summary: test detail"),
			other:    function.NewFunctionError("test summary: test detail"),
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
