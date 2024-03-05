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

func TestFunctionError_Equal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErr  *function.FuncError
		other    *function.FuncError
		expected bool
	}{
		"nil-nil": {
			expected: true,
		},
		"empty-nil": {
			funcErr:  &function.FuncError{},
			other:    nil,
			expected: false,
		},
		"nil-empty": {
			funcErr:  nil,
			other:    &function.FuncError{},
			expected: false,
		},
		"error-nil": {
			funcErr:  function.NewFuncError("test summary: test detail"),
			other:    nil,
			expected: false,
		},
		"nil-error": {
			funcErr:  nil,
			other:    function.NewFuncError("test summary: test detail"),
			expected: false,
		},
		"different-text": {
			funcErr:  function.NewFuncError("test summary: test detail"),
			other:    function.NewFuncError("test summary: different detail"),
			expected: false,
		},
		"different-type": {
			funcErr:  function.NewFuncError("test summary: test detail"),
			other:    function.NewArgumentFuncError(int64(0), "test summary: test detail"),
			expected: false,
		},
		"different-argument": {
			funcErr:  function.NewArgumentFuncError(int64(0), "test summary: test detail"),
			other:    function.NewArgumentFuncError(int64(1), "test summary: test detail"),
			expected: false,
		},
		"matching-text": {
			funcErr:  function.NewFuncError("test summary: test detail"),
			other:    function.NewFuncError("test summary: test detail"),
			expected: true,
		},
		"matching-argument": {
			funcErr:  function.NewArgumentFuncError(int64(0), "test summary: test detail"),
			other:    function.NewArgumentFuncError(int64(0), "test summary: test detail"),
			expected: true,
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

func TestConcatFuncErrors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		funcErr  *function.FuncError
		other    *function.FuncError
		expected *function.FuncError
	}{
		"nil-nil": {},
		"empty-nil-slice": {
			funcErr: &function.FuncError{},
			other:   nil,
		},
		"empty-empty": {
			funcErr: &function.FuncError{},
			other:   &function.FuncError{},
		},
		"text-nil": {
			funcErr: &function.FuncError{
				Text: "function error one",
			},
			other: nil,
			expected: &function.FuncError{
				Text: "function error one",
			},
		},
		"text-empty": {
			funcErr: &function.FuncError{
				Text: "function error one",
			},
			other: &function.FuncError{},
			expected: &function.FuncError{
				Text: "function error one",
			},
		},
		"nil-text": {
			funcErr: nil,
			other: &function.FuncError{
				Text: "function error two",
			},
			expected: &function.FuncError{
				Text: "function error two",
			},
		},
		"empty-text": {
			funcErr: &function.FuncError{},
			other: &function.FuncError{
				Text: "function error two",
			},
			expected: &function.FuncError{
				Text: "function error two",
			},
		},
		"text-text": {
			funcErr: &function.FuncError{
				Text: "function error one",
			},
			other: &function.FuncError{
				Text: "function error two",
			},
			expected: &function.FuncError{
				Text: "function error one\nfunction error two",
			},
		},
		"nil-argument": {
			funcErr: nil,
			other: &function.FuncError{
				FunctionArgument: pointer(int64(0)),
			},
			expected: &function.FuncError{
				FunctionArgument: pointer(int64(0)),
			},
		},
		"argument-nil": {
			funcErr: &function.FuncError{
				FunctionArgument: pointer(int64(0)),
			},
			other: nil,
			expected: &function.FuncError{
				FunctionArgument: pointer(int64(0)),
			},
		},
		"argument-precedence": {
			funcErr: &function.FuncError{
				FunctionArgument: pointer(int64(0)),
			},
			other: &function.FuncError{
				FunctionArgument: pointer(int64(1)),
			},
			expected: &function.FuncError{
				FunctionArgument: pointer(int64(0)),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.ConcatFuncErrors(tc.funcErr, tc.other)

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFuncErrorFromDiags(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags       diag.Diagnostics
		expected    *function.FuncError
		expectedLog []map[string]interface{}
	}{
		"nil": {},
		"empty": {
			diags: diag.Diagnostics{},
		},
		"error": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
			},
			expected: &function.FuncError{
				Text: "one summary: one detail",
			},
		},
		"warning": {
			diags: diag.Diagnostics{
				diag.NewWarningDiagnostic("one summary", "one detail"),
			},
			expectedLog: []map[string]interface{}{
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "one detail",
					"summary":  "one summary",
				},
			},
		},
		"error-warning": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: function.NewFuncError("one summary: one detail"),
			expectedLog: []map[string]interface{}{
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "two detail",
					"summary":  "two summary",
				},
			},
		},
		"multiple": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
				diag.NewErrorDiagnostic("three summary", "three detail"),
				diag.NewWarningDiagnostic("four summary", "four detail"),
			},
			expected: function.NewFuncError("one summary: one detail\nthree summary: three detail"),
			expectedLog: []map[string]interface{}{
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "two detail",
					"summary":  "two summary",
				},
				{
					"@level":   "warn",
					"@message": "warning: call function",
					"@module":  "provider",
					"detail":   "four detail",
					"summary":  "four summary",
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			ctx := tflogtest.RootLogger(context.Background(), &output)

			got := function.FuncErrorFromDiags(ctx, tc.diags)

			entries, err := tflogtest.MultilineJSONDecode(&output)

			if err != nil {
				t.Fatalf("unable to read multiple line JSON: %s", err)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(entries, tc.expectedLog); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
