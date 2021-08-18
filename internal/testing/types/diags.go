package types

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	TestErrorDiagnostic = diag.NewErrorDiagnostic(
		"Error Diagnostic",
		"This is an error.",
	)
	TestWarningDiagnostic = diag.NewWarningDiagnostic(
		"Warning Diagnostic",
		"This is a warning.",
	)
)
