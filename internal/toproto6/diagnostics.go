package toproto6

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Diagnostics converts the diagnostics into the tfprotov6 collection type.
func Diagnostics(diagnostics diag.Diagnostics) []*tfprotov6.Diagnostic {
	var results []*tfprotov6.Diagnostic

	for _, diagnostic := range diagnostics {
		tfprotov6Diagnostic := &tfprotov6.Diagnostic{
			Detail:   diagnostic.Detail(),
			Severity: diagnostic.Severity().ToTfprotov6DiagnosticSeverity(),
			Summary:  diagnostic.Summary(),
		}

		if diagWithPath, ok := diagnostic.(diag.DiagnosticWithPath); ok {
			tfprotov6Diagnostic.Attribute = diagWithPath.Path()
		}

		results = append(results, tfprotov6Diagnostic)
	}

	return results
}
