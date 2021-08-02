package types

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	TestErrorDiagnostic = &tfprotov6.Diagnostic{
		Severity: tfprotov6.DiagnosticSeverityError,
		Summary:  "Error Diagnostic",
		Detail:   "This is an error.",
	}
	TestWarningDiagnostic = &tfprotov6.Diagnostic{
		Severity: tfprotov6.DiagnosticSeverityWarning,
		Summary:  "Warning Diagnostic",
		Detail:   "This is a warning.",
	}
)
