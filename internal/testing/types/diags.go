package types

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestErrorDiagnostic(path *tftypes.AttributePath) diag.DiagnosticWithPath {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Error Diagnostic",
		"This is an error.",
	)
}

func TestWarningDiagnostic(path *tftypes.AttributePath) diag.DiagnosticWithPath {
	return diag.NewAttributeWarningDiagnostic(
		path,
		"Warning Diagnostic",
		"This is a warning.",
	)
}
