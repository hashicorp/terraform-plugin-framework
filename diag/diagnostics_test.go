// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func TestDiagnosticsAddAttributeError(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		path     path.Path
		summary  string
		detail   string
		expected diag.Diagnostics
	}{
		"nil-add": {
			diags:   nil,
			path:    path.Root("test"),
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
			},
		},
		"add": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
			},
			path:    path.Root("test"),
			summary: "three summary",
			detail:  "three detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "three summary", "three detail"),
			},
		},
		"duplicate": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
			},
			path:    path.Root("test"),
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.diags.AddAttributeError(tc.path, tc.summary, tc.detail)

			if diff := cmp.Diff(tc.diags, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestDiagnosticsAddAttributeWarning(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		path     path.Path
		summary  string
		detail   string
		expected diag.Diagnostics
	}{
		"nil-add": {
			diags:   nil,
			path:    path.Root("test"),
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "one summary", "one detail"),
			},
		},
		"add": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
			},
			path:    path.Root("test"),
			summary: "three summary",
			detail:  "three detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "three summary", "three detail"),
			},
		},
		"duplicate": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
			},
			path:    path.Root("test"),
			summary: "two summary",
			detail:  "two detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.diags.AddAttributeWarning(tc.path, tc.summary, tc.detail)

			if diff := cmp.Diff(tc.diags, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestDiagnosticsAddError(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		summary  string
		detail   string
		expected diag.Diagnostics
	}{
		"nil-add": {
			diags:   nil,
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
			},
		},
		"add": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			summary: "three summary",
			detail:  "three detail",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
				diag.NewErrorDiagnostic("three summary", "three detail"),
			},
		},
		"duplicate": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.diags.AddError(tc.summary, tc.detail)

			if diff := cmp.Diff(tc.diags, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestDiagnosticsAddWarning(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		summary  string
		detail   string
		expected diag.Diagnostics
	}{
		"nil-add": {
			diags:   nil,
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewWarningDiagnostic("one summary", "one detail"),
			},
		},
		"add": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			summary: "three summary",
			detail:  "three detail",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
				diag.NewWarningDiagnostic("three summary", "three detail"),
			},
		},
		"duplicate": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			summary: "two summary",
			detail:  "two detail",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.diags.AddWarning(tc.summary, tc.detail)

			if diff := cmp.Diff(tc.diags, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestDiagnosticsAppend(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		in       diag.Diagnostics
		expected diag.Diagnostics
	}{
		"nil-append": {
			diags: nil,
			in: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
		},
		"append": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in: diag.Diagnostics{
				diag.NewErrorDiagnostic("three summary", "three detail"),
				diag.NewWarningDiagnostic("four summary", "four detail"),
			},
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
				diag.NewErrorDiagnostic("three summary", "three detail"),
				diag.NewWarningDiagnostic("four summary", "four detail"),
			},
		},
		"append-less-specific": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			},
			in: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
		},
		"append-more-specific": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			},
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
				diag.NewAttributeErrorDiagnostic(path.Root("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			},
		},
		"empty-diagnostics": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in: nil,
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
		},
		"empty-diagnostics-elements": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in: diag.Diagnostics{
				nil,
				nil,
			},
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
		},
		"duplicate": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.diags.Append(tc.in...)

			if diff := cmp.Diff(tc.diags, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestDiagnosticsContains(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		in       diag.Diagnostic
		expected bool
	}{
		"matching-basic": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in:       diag.NewWarningDiagnostic("two summary", "two detail"),
			expected: true,
		},
		"matching-attribute-path": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			},
			in:       diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			expected: true,
		},
		"nil-diagnostics": {
			diags:    nil,
			in:       diag.NewErrorDiagnostic("one summary", "one detail"),
			expected: false,
		},
		"nil-in": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in:       nil,
			expected: false,
		},
		"different-attribute-path": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			},
			in:       diag.NewAttributeWarningDiagnostic(path.Root("different"), "two summary", "two detail"),
			expected: false,
		},
		"different-detail": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in:       diag.NewWarningDiagnostic("two summary", "different detail"),
			expected: false,
		},
		"different-severity": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in:       diag.NewWarningDiagnostic("one summary", "one detail"),
			expected: false,
		},
		"different-summary": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in:       diag.NewWarningDiagnostic("different summary", "two detail"),
			expected: false,
		},
		"different-type-less-specific": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			},
			in:       diag.NewWarningDiagnostic("two summary", "two detail"),
			expected: false,
		},
		"different-type-more-specific": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in:       diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			expected: false,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.diags.Contains(tc.in)

			if got != tc.expected {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}
		})
	}
}

func TestDiagnosticsEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diagnostics diag.Diagnostics
		other       diag.Diagnostics
		expected    bool
	}{
		"nil-nil": {
			diagnostics: nil,
			other:       nil,
			expected:    true,
		},
		"nil-empty": {
			diagnostics: nil,
			other:       diag.Diagnostics{},
			expected:    true,
		},
		"empty-nil": {
			diagnostics: diag.Diagnostics{},
			other:       nil,
			expected:    true,
		},
		"empty-empty": {
			diagnostics: diag.Diagnostics{},
			other:       diag.Diagnostics{},
			expected:    true,
		},
		"different-length": {
			diagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewErrorDiagnostic("two summary", "two detail"),
			},
			other: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
			},
			expected: false,
		},
		"Attribute-different": {
			diagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "test summary", "test detail"),
			},
			other: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("not-test"), "test summary", "test detail"),
			},
			expected: false,
		},
		"Attribute-equal": {
			diagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "test summary", "test detail"),
			},
			other: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("test"), "test summary", "test detail"),
			},
			expected: true,
		},
		"Detail-different": {
			diagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			other: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "not test detail"),
			},
			expected: false,
		},
		"Detail-equal": {
			diagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			other: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			expected: true,
		},
		"Severity-different": {
			diagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			other: diag.Diagnostics{
				diag.NewWarningDiagnostic("test summary", "test detail"),
			},
			expected: false,
		},
		"Severity-equal": {
			diagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			other: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			expected: true,
		},
		"Summary-different": {
			diagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			other: diag.Diagnostics{
				diag.NewErrorDiagnostic("not test summary", "test detail"),
			},
			expected: false,
		},
		"Summary-equal": {
			diagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			other: diag.Diagnostics{
				diag.NewErrorDiagnostic("test summary", "test detail"),
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.diagnostics.Equal(testCase.other)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestDiagnosticsHasError(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		expected bool
	}{
		"matching-basic": {
			diags: diag.Diagnostics{
				diag.NewWarningDiagnostic("one summary", "one detail"),
				diag.NewErrorDiagnostic("two summary", "two detail"),
			},
			expected: true,
		},
		"matching-attribute-path": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Root("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("warning"), "two summary", "two detail"),
			},
			expected: true,
		},
		"nil-diagnostics": {
			diags:    nil,
			expected: false,
		},
		"different-severity": {
			diags: diag.Diagnostics{
				diag.NewWarningDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: false,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.diags.HasError()

			if got != tc.expected {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}
		})
	}
}

func TestDiagnosticsErrorsCount(t *testing.T) {
	t.Parallel()

	type testCase struct {
		diags    diag.Diagnostics
		expected int
	}
	tests := map[string]testCase{
		"nil": {
			diags:    nil,
			expected: 0,
		},
		"empty": {
			diags:    diag.Diagnostics{},
			expected: 0,
		},
		"errors": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Summary", "Error detail."),
				diag.NewWarningDiagnostic("Warning Summary", "Warning detail."),
			},
			expected: 1,
		},
		"warnings": {
			diags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Error Summary", "Error detail."),
			},
			expected: 0,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.diags.ErrorsCount()

			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Fatalf("expected: %q, got: %q", test.expected, got)
			}
		})
	}
}

func TestDiagnosticsWarningsCount(t *testing.T) {
	t.Parallel()

	type testCase struct {
		diags    diag.Diagnostics
		expected int
	}
	tests := map[string]testCase{
		"nil": {
			diags:    nil,
			expected: 0,
		},
		"empty": {
			diags:    diag.Diagnostics{},
			expected: 0,
		},
		"errors": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Summary", "Error detail."),
				diag.NewWarningDiagnostic("Warning Summary", "Warning detail."),
			},
			expected: 1,
		},
		"warnings": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Summary", "Error detail."),
			},
			expected: 0,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.diags.WarningsCount()

			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Fatalf("expected: %q, got: %q", test.expected, got)
			}
		})
	}
}

func TestDiagnosticsErrors(t *testing.T) {
	t.Parallel()

	type testCase struct {
		diags    diag.Diagnostics
		expected diag.Diagnostics
	}
	tests := map[string]testCase{
		"nil": {
			diags:    nil,
			expected: diag.Diagnostics{},
		},
		"empty": {
			diags:    diag.Diagnostics{},
			expected: diag.Diagnostics{},
		},
		"errors": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Summary", "Error detail."),
				diag.NewWarningDiagnostic("Warning Summary", "Warning detail."),
			},
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Summary", "Error detail."),
			},
		},
		"warnings": {
			diags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Warning Summary", "Warning detail."),
			},
			expected: diag.Diagnostics{},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.diags.Errors()

			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Fatalf("expected: %q, got: %q", test.expected, got)
			}
		})
	}
}

func TestDiagnosticsWarnings(t *testing.T) {
	t.Parallel()

	type testCase struct {
		diags    diag.Diagnostics
		expected diag.Diagnostics
	}
	tests := map[string]testCase{
		"nil": {
			diags:    nil,
			expected: diag.Diagnostics{},
		},
		"empty": {
			diags:    diag.Diagnostics{},
			expected: diag.Diagnostics{},
		},
		"errors": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Summary", "Error detail."),
			},
			expected: diag.Diagnostics{},
		},
		"warnings": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Summary", "Error detail."),
				diag.NewWarningDiagnostic("Warning Summary", "Warning detail."),
			},
			expected: diag.Diagnostics{
				diag.NewWarningDiagnostic("Warning Summary", "Warning detail."),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.diags.Warnings()

			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Fatalf("expected: %q, got: %q", test.expected, got)
			}
		})
	}
}
