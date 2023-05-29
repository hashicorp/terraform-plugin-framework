// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDiagnosticSeverity(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		severity diag.Severity
		expected tfprotov5.DiagnosticSeverity
	}{
		"error": {
			severity: diag.SeverityError,
			expected: tfprotov5.DiagnosticSeverityError,
		},
		"invalid": {
			severity: diag.SeverityInvalid,
			expected: tfprotov5.DiagnosticSeverityInvalid,
		},
		"warning": {
			severity: diag.SeverityWarning,
			expected: tfprotov5.DiagnosticSeverityWarning,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.DiagnosticSeverity(testCase.severity)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestDiagnostics(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diags    diag.Diagnostics
		expected []*tfprotov5.Diagnostic
	}{
		"nil": {
			diags:    nil,
			expected: nil,
		},
		"Diagnostic-SeverityInvalid": {
			diags: diag.Diagnostics{
				invalidSeverityDiagnostic{},
			},
			expected: []*tfprotov5.Diagnostic{
				{
					Detail:   invalidSeverityDiagnostic{}.Detail(),
					Severity: tfprotov5.DiagnosticSeverityInvalid,
					Summary:  invalidSeverityDiagnostic{}.Summary(),
				},
			},
		},
		"Diagnostic": {
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("one summary", "one detail"),
				diag.NewWarningDiagnostic("two summary", "two detail"),
			},
			expected: []*tfprotov5.Diagnostic{
				{
					Detail:   "one detail",
					Severity: tfprotov5.DiagnosticSeverityError,
					Summary:  "one summary",
				},
				{
					Detail:   "two detail",
					Severity: tfprotov5.DiagnosticSeverityWarning,
					Summary:  "two summary",
				},
			},
		},
		"DiagnosticWithPath": {
			diags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(path.Empty(), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(path.Root("test"), "two summary", "two detail"),
			},
			expected: []*tfprotov5.Diagnostic{
				{
					Attribute: tftypes.NewAttributePath(),
					Detail:    "one detail",
					Severity:  tfprotov5.DiagnosticSeverityError,
					Summary:   "one summary",
				},
				{
					Attribute: tftypes.NewAttributePath().WithAttributeName("test"),
					Detail:    "two detail",
					Severity:  tfprotov5.DiagnosticSeverityWarning,
					Summary:   "two summary",
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto5.Diagnostics(context.Background(), tc.diags)

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

var _ diag.Diagnostic = invalidSeverityDiagnostic{}

type invalidSeverityDiagnostic struct{}

func (d invalidSeverityDiagnostic) Detail() string {
	return "detail for invalid severity diagnostic"
}

func (d invalidSeverityDiagnostic) Equal(other diag.Diagnostic) bool {
	isd, ok := other.(invalidSeverityDiagnostic)

	if !ok {
		return false
	}

	return isd.Summary() == d.Summary() && isd.Detail() == d.Detail() && isd.Severity() == d.Severity()
}

func (d invalidSeverityDiagnostic) Severity() diag.Severity {
	return diag.SeverityInvalid
}

func (d invalidSeverityDiagnostic) Summary() string {
	return "summary for invalid severity diagnostic"
}
