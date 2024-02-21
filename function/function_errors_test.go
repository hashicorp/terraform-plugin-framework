// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

func TestFunctionErrors_AddArgumentError(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErrs function.FunctionErrors
		position int
		msg      string
		expected function.FunctionErrors
	}{
		"nil-add": {
			funcErrs: nil,
			position: 0,
			msg:      "one summary: one detail",
			expected: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
			},
		},
		"add": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
				function.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			position: 0,
			msg:      "three summary: three detail",
			expected: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
				function.NewArgumentFunctionError(0, "two summary: two detail"),
				function.NewArgumentFunctionError(0, "three summary: three detail"),
			},
		},
		"duplicate": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
				function.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			position: 0,
			msg:      "one summary: one detail",
			expected: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
				function.NewArgumentFunctionError(0, "two summary: two detail"),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.funcErrs.AddArgumentError(tc.position, tc.msg)

			if diff := cmp.Diff(tc.funcErrs, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFunctionErrors_AddError(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErrs function.FunctionErrors
		msg      string
		expected function.FunctionErrors
	}{
		"nil-add": {
			funcErrs: nil,
			msg:      "one summary: one detail",
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
		},
		"add": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
			msg: "three summary: three detail",
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
				function.NewFunctionError("three summary: three detail"),
			},
		},
		"duplicate": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
			msg: "one summary: one detail",
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.funcErrs.AddError(tc.msg)

			if diff := cmp.Diff(tc.funcErrs, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFunctionErrors_Append(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErrs function.FunctionErrors
		in       function.FunctionErrors
		expected function.FunctionErrors
	}{
		"nil-append": {
			funcErrs: nil,
			in: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
		},
		"append": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
			in: function.FunctionErrors{
				function.NewFunctionError("three summary: three detail"),
				function.NewFunctionError("four summary: four detail"),
			},
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
				function.NewFunctionError("three summary: three detail"),
				function.NewFunctionError("four summary: four detail"),
			},
		},
		"append-less-specific": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			in: function.FunctionErrors{
				function.NewFunctionError("two summary: two detail"),
			},
			expected: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
		},
		"append-more-specific": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			in: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewArgumentFunctionError(0, "two summary: two detail"),
			},
		},
		"empty-function-errors": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			in: nil,
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
		},
		"empty-function-errors-elements": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			in: function.FunctionErrors{
				nil,
				nil,
			},
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
		},
		"duplicate": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			in: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.funcErrs.Append(tc.in...)

			if diff := cmp.Diff(tc.funcErrs, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFunctionErrors_Contains(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErrs function.FunctionErrors
		in       function.FunctionError
		expected bool
	}{
		"matching-basic": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			in:       function.NewFunctionError("one summary: one detail"),
			expected: true,
		},
		"matching-function-argument": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			in:       function.NewArgumentFunctionError(0, "one summary: one detail"),
			expected: true,
		},
		"nil-function-errors": {
			funcErrs: nil,
			in:       function.NewFunctionError("one summary: one detail"),
			expected: false,
		},
		"nil-in": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			in:       nil,
			expected: false,
		},
		"different-function-argument": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			in:       function.NewArgumentFunctionError(1, "one summary: one detail"),
			expected: false,
		},
		"different-msg": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			in:       function.NewFunctionError("one summary: different detail"),
			expected: false,
		},
		"different-type-less-specific": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			in:       function.NewFunctionError("one summary: one detail"),
			expected: false,
		},
		"different-type-more-specific": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			in:       function.NewArgumentFunctionError(0, "one summary: one detail"),
			expected: false,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.funcErrs.Contains(tc.in)

			if got != tc.expected {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}
		})
	}
}

func TestFunctionErrors_Equal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErrs function.FunctionErrors
		other    function.FunctionErrors
		expected bool
	}{
		"nil-nil": {
			funcErrs: nil,
			other:    nil,
			expected: true,
		},
		"nil-empty": {
			funcErrs: nil,
			other:    function.FunctionErrors{},
			expected: true,
		},
		"empty-nil": {
			funcErrs: function.FunctionErrors{},
			other:    nil,
			expected: true,
		},
		"empty-empty": {
			funcErrs: function.FunctionErrors{},
			other:    function.FunctionErrors{},
			expected: true,
		},
		"different-length": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
			other: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			expected: false,
		},
		"function-argument-different": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			other: function.FunctionErrors{
				function.NewArgumentFunctionError(1, "one summary: one detail"),
			},
			expected: false,
		},
		"function-argument-equal": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			other: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			expected: true,
		},
		"msg-different": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			other: function.FunctionErrors{
				function.NewFunctionError("one summary: different detail"),
			},
			expected: false,
		},
		"msg-equal": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			other: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.funcErrs.Equal(&testCase.other)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestFunctionErrors_Error(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErrs function.FunctionErrors
		expected string
	}{
		"nil": {
			funcErrs: nil,
			expected: "",
		},
		"empty": {
			funcErrs: function.FunctionErrors{},
			expected: "",
		},
		"same-type-basic": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
			expected: "one summary: one detail\ntwo summary: two detail\n",
		},
		"same-type-function-argument": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
				function.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			expected: "one summary: one detail\ntwo summary: two detail\n",
		},
		"different-type-less-specific": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
			expected: "one summary: one detail\ntwo summary: two detail\n",
		},
		"different-type-more-specific": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			expected: "one summary: one detail\ntwo summary: two detail\n",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.funcErrs.Error()

			if got != testCase.expected {
				t.Errorf("expected %s, got %s", testCase.expected, got)
			}
		})
	}
}

func TestFunctionErrors_HasError(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErrs function.FunctionErrors
		expected bool
	}{
		"matching-basic": {
			funcErrs: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
				function.NewFunctionError("two summary: two detail"),
			},
			expected: true,
		},
		"matching-function-argument": {
			funcErrs: function.FunctionErrors{
				function.NewArgumentFunctionError(0, "one summary: one detail"),
				function.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			expected: true,
		},
		"nil-function-errors": {
			funcErrs: nil,
			expected: false,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.funcErrs.HasError()

			if got != tc.expected {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}
		})
	}
}

func TestFunctionErrorsFromDiags(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tflogtest.RootLogger(context.Background(), &output)

	testCases := map[string]struct {
		diags       diag.Diagnostics
		expected    function.FunctionErrors
		expectedLog string
	}{
		"log": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: function.FunctionErrors{
				function.NewFunctionError("one summary: one detail"),
			},
			expectedLog: `{"@level":"warn","@message":"warning: call function","@module":"provider","detail":"two detail","summary":"two summary"}` + "\n",
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.FunctionErrorsFromDiags(ctx, tc.diags)

			if !got.Equal(&tc.expected) {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}

			if output.String() != tc.expectedLog {
				t.Errorf("Unexpected log: got: %s, wanted: %s", output.String(), tc.expectedLog)
			}
		})
	}
}
