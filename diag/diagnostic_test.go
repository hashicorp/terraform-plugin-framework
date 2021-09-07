package diag_test

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

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
