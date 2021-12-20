package diag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attrpath"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDiagnosticsAddAttributeError(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		path     attrpath.Path
		summary  string
		detail   string
		expected diag.Diagnostics
	}{
		"nil-add": {
			diags:   nil,
			path:    attrpath.New().Attribute("test"),
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
			},
		},
		"add": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
			},
			path:    attrpath.New().Attribute("test"),
			summary: "three summary",
			detail:  "three detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "three summary", "three detail"),
			},
		},
		"duplicate": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
			},
			path:    attrpath.New().Attribute("test"),
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
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
		path     attrpath.Path
		summary  string
		detail   string
		expected diag.Diagnostics
	}{
		"nil-add": {
			diags:   nil,
			path:    attrpath.New().Attribute("test"),
			summary: "one summary",
			detail:  "one detail",
			expected: diag.Diagnostics{
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
			},
		},
		"add": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
			},
			path:    attrpath.New().Attribute("test"),
			summary: "three summary",
			detail:  "three detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "three summary", "three detail"),
			},
		},
		"duplicate": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
			},
			path:    attrpath.New().Attribute("test"),
			summary: "two summary",
			detail:  "two detail",
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("test"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
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
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
			},
			in: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
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
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
			},
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
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
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
			},
			in:       diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
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
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
			},
			in:       diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("different"), "two summary", "two detail"),
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
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
			},
			in:       diag.NewWarningDiagnostic("two summary", "two detail"),
			expected: false,
		},
		"different-type-more-specific": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			in:       diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
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
				diag.NewAttributeErrorDiagnostic(attrpath.New().Attribute("error"), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("warning"), "two summary", "two detail"),
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

func TestDiagnosticsToTfprotov6Diagnostics(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		expected []*tfprotov6.Diagnostic
	}{
		"nil": {
			diags:    nil,
			expected: nil,
		},
		"Diagnostic-SeverityInvalid": {
			diags: diag.Diagnostics{
				invalidSeverityDiagnostic{},
			},
			expected: []*tfprotov6.Diagnostic{
				{
					Detail:   invalidSeverityDiagnostic{}.Detail(),
					Severity: tfprotov6.DiagnosticSeverityInvalid,
					Summary:  invalidSeverityDiagnostic{}.Summary(),
				},
			},
		},
		"Diagnostic": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: []*tfprotov6.Diagnostic{
				{
					Detail:   "one detail",
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "one summary",
				},
				{
					Detail:   "two detail",
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "two summary",
				},
			},
		},
		"DiagnosticWithPath": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(attrpath.New(), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(attrpath.New().Attribute("test"), "two summary", "two detail"),
			},
			expected: []*tfprotov6.Diagnostic{
				{
					Attribute: tftypes.NewAttributePath(),
					Detail:    "one detail",
					Severity:  tfprotov6.DiagnosticSeverityError,
					Summary:   "one summary",
				},
				{
					Attribute: tftypes.NewAttributePath().WithAttributeName("test"),
					Detail:    "two detail",
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Summary:   "two summary",
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.diags.ToTfprotov6Diagnostics()

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}
