package toproto5

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// DiagnosticSeverity converts diag.Severity into tfprotov5.DiagnosticSeverity.
func DiagnosticSeverity(s diag.Severity) tfprotov5.DiagnosticSeverity {
	switch s {
	case diag.SeverityError:
		return tfprotov5.DiagnosticSeverityError
	case diag.SeverityWarning:
		return tfprotov5.DiagnosticSeverityWarning
	default:
		return tfprotov5.DiagnosticSeverityInvalid
	}
}

// Diagnostics converts the diagnostics into the tfprotov5 collection type.
func Diagnostics(diagnostics diag.Diagnostics) []*tfprotov5.Diagnostic {
	var results []*tfprotov5.Diagnostic

	for _, diagnostic := range diagnostics {
		tfprotov5Diagnostic := &tfprotov5.Diagnostic{
			Detail:   diagnostic.Detail(),
			Severity: DiagnosticSeverity(diagnostic.Severity()),
			Summary:  diagnostic.Summary(),
		}

		if diagWithPath, ok := diagnostic.(diag.DiagnosticWithPath); ok {
			tfprotov5Diagnostic.Attribute = diagWithPath.Path()
		}

		results = append(results, tfprotov5Diagnostic)
	}

	return results
}
