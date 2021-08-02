package reflect_test

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TODO: Replace with diagnostics abstraction
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/24
func diagsContainsDetail(in []*tfprotov6.Diagnostic, detail string) bool {
	for _, diag := range in {
		if diag == nil {
			continue
		}

		if strings.Contains(diag.Detail, detail) {
			return true
		}
	}
	return false
}

// TODO: Replace with diagnostics abstraction
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/24
func diagsHasErrors(in []*tfprotov6.Diagnostic) bool {
	for _, diag := range in {
		if diag == nil {
			continue
		}

		if diag.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

// TODO: Replace with diagnostics abstraction
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/24
func diagsString(in []*tfprotov6.Diagnostic) string {
	var b strings.Builder

	for _, diag := range in {
		if diag == nil {
			continue
		}

		// Diagnostic does not have .String() method
		b.WriteString("\n")
		b.WriteString(diagString(diag))
		b.WriteString("\n")
	}

	return b.String()
}

// TODO: Replace with diagnostics abstraction
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/24
func diagString(diag *tfprotov6.Diagnostic) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Severity: %s\n", diag.Severity))
	b.WriteString(fmt.Sprintf("Summary: %s\n", diag.Summary))
	b.WriteString(fmt.Sprintf("Detail: %s\n", diag.Detail))
	if diag.Attribute != nil {
		b.WriteString(fmt.Sprintf("Attribute: %s\n", diag.Attribute))
	}

	return b.String()
}
