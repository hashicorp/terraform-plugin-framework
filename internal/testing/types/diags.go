package types

import (
	"github.com/hashicorp/terraform-plugin-framework/attrpath"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestErrorDiagnostic(path attrpath.Path) diag.DiagnosticWithPath {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Error Diagnostic",
		"This is an error.",
	)
}

func TestWarningDiagnostic(path attrpath.Path) diag.DiagnosticWithPath {
	return diag.NewAttributeWarningDiagnostic(
		path,
		"Warning Diagnostic",
		"This is a warning.",
	)
}
