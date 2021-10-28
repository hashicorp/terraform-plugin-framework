package diag_test

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ diag.Diagnostic = &invalidSeverityDiagnostic{}

type invalidSeverityDiagnostic struct{}

func (d *invalidSeverityDiagnostic) Detail() string {
	return "detail for invalid severity diagnostic"
}

func (d *invalidSeverityDiagnostic) Equal(other diag.Diagnostic) bool {
	if d == nil && other == nil {
		return true
	}

	isd, ok := other.(*invalidSeverityDiagnostic)

	if !ok {
		return false
	}

	return isd.Summary() == d.Summary() && isd.Detail() == d.Detail() && isd.Severity() == d.Severity() && isd.Path().Equal(d.Path())
}

func (d *invalidSeverityDiagnostic) Path() *tftypes.AttributePath {
	return nil
}

func (d *invalidSeverityDiagnostic) SetPath(path *tftypes.AttributePath) {}

func (d *invalidSeverityDiagnostic) Severity() diag.Severity {
	return diag.SeverityInvalid
}

func (d *invalidSeverityDiagnostic) Summary() string {
	return "summary for invalid severity diagnostic"
}
