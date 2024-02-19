// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/fwerror"
)

func TestFunctionErrors_AddArgumentError(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErrs fwerror.FunctionErrors
		position int
		msg      string
		expected fwerror.FunctionErrors
	}{
		"nil-add": {
			funcErrs: nil,
			position: 0,
			msg:      "one summary: one detail",
			expected: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			},
		},
		"add": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			position: 0,
			msg:      "three summary: three detail",
			expected: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
				fwerror.NewArgumentFunctionError(0, "three summary: three detail"),
			},
		},
		"duplicate": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			position: 0,
			msg:      "one summary: one detail",
			expected: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
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
		funcErrs fwerror.FunctionErrors
		msg      string
		expected fwerror.FunctionErrors
	}{
		"nil-add": {
			funcErrs: nil,
			msg:      "one summary: one detail",
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
		},
		"add": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
			msg: "three summary: three detail",
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
				fwerror.NewFunctionError("three summary: three detail"),
			},
		},
		"duplicate": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
			msg: "one summary: one detail",
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
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
		funcErrs fwerror.FunctionErrors
		in       fwerror.FunctionErrors
		expected fwerror.FunctionErrors
	}{
		"nil-append": {
			funcErrs: nil,
			in: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
		},
		"append": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
			in: fwerror.FunctionErrors{
				fwerror.NewFunctionError("three summary: three detail"),
				fwerror.NewFunctionError("four summary: four detail"),
			},
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
				fwerror.NewFunctionError("three summary: three detail"),
				fwerror.NewFunctionError("four summary: four detail"),
			},
		},
		"append-less-specific": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			in: fwerror.FunctionErrors{
				fwerror.NewFunctionError("two summary: two detail"),
			},
			expected: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
		},
		"append-more-specific": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			in: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
			},
		},
		"empty-function-errors": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			in: nil,
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
		},
		"empty-function-errors-elements": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			in: fwerror.FunctionErrors{
				nil,
				nil,
			},
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
		},
		"duplicate": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			in: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
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
		funcErrs fwerror.FunctionErrors
		in       fwerror.FunctionError
		expected bool
	}{
		"matching-basic": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			in:       fwerror.NewFunctionError("one summary: one detail"),
			expected: true,
		},
		"matching-function-argument": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			in:       fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			expected: true,
		},
		"nil-function-errors": {
			funcErrs: nil,
			in:       fwerror.NewFunctionError("one summary: one detail"),
			expected: false,
		},
		"nil-in": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			in:       nil,
			expected: false,
		},
		"different-function-argument": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			in:       fwerror.NewArgumentFunctionError(1, "one summary: one detail"),
			expected: false,
		},
		"different-msg": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			in:       fwerror.NewFunctionError("one summary: different detail"),
			expected: false,
		},
		"different-type-less-specific": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			in:       fwerror.NewFunctionError("one summary: one detail"),
			expected: false,
		},
		"different-type-more-specific": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			in:       fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
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
		funcErrs fwerror.FunctionErrors
		other    fwerror.FunctionErrors
		expected bool
	}{
		"nil-nil": {
			funcErrs: nil,
			other:    nil,
			expected: true,
		},
		"nil-empty": {
			funcErrs: nil,
			other:    fwerror.FunctionErrors{},
			expected: true,
		},
		"empty-nil": {
			funcErrs: fwerror.FunctionErrors{},
			other:    nil,
			expected: true,
		},
		"empty-empty": {
			funcErrs: fwerror.FunctionErrors{},
			other:    fwerror.FunctionErrors{},
			expected: true,
		},
		"different-length": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
			other: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			expected: false,
		},
		"function-argument-different": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			other: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(1, "one summary: one detail"),
			},
			expected: false,
		},
		"function-argument-equal": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			other: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
			},
			expected: true,
		},
		"msg-different": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			other: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: different detail"),
			},
			expected: false,
		},
		"msg-equal": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			other: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
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
		funcErrs fwerror.FunctionErrors
		expected string
	}{
		"nil": {
			funcErrs: nil,
			expected: "",
		},
		"empty": {
			funcErrs: fwerror.FunctionErrors{},
			expected: "",
		},
		"same-type-basic": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
			expected: "one summary: one detail\ntwo summary: two detail\n",
		},
		"same-type-function-argument": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
			},
			expected: "one summary: one detail\ntwo summary: two detail\n",
		},
		"different-type-less-specific": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
			expected: "one summary: one detail\ntwo summary: two detail\n",
		},
		"different-type-more-specific": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
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
		funcErrs fwerror.FunctionErrors
		expected bool
	}{
		"matching-basic": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
				fwerror.NewFunctionError("two summary: two detail"),
			},
			expected: true,
		},
		"matching-function-argument": {
			funcErrs: fwerror.FunctionErrors{
				fwerror.NewArgumentFunctionError(0, "one summary: one detail"),
				fwerror.NewArgumentFunctionError(0, "two summary: two detail"),
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
		expected    fwerror.FunctionErrors
		expectedLog string
	}{
		"log": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: fwerror.FunctionErrors{
				fwerror.NewFunctionError("one summary: one detail"),
			},
			expectedLog: `{"@level":"warn","@message":"warning: call function","@module":"provider","detail":"two detail","summary":"two summary"}` + "\n",
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwerror.FunctionErrorsFromDiags(ctx, tc.diags)

			if !got.Equal(&tc.expected) {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}

			if output.String() != tc.expectedLog {
				t.Errorf("Unexpected log: got: %s, wanted: %s", output.String(), tc.expectedLog)
			}
		})
	}
}
