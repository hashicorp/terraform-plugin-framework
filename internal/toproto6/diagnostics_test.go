package toproto6_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDiagnostics(t *testing.T) {
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
				diag.NewAttributeErrorDiagnostic(tftypes.NewAttributePath(), "one summary", "one detail"),
				diag.NewAttributeWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"), "two summary", "two detail"),
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

			got := toproto6.Diagnostics(tc.diags)

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
