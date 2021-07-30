package tfsdk

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

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
		b.WriteString(fmt.Sprintf("Severity: %s\n", diag.Severity))
		b.WriteString(fmt.Sprintf("Summary: %s\n", diag.Summary))
		b.WriteString(fmt.Sprintf("Detail: %s\n", diag.Detail))
		if diag.Attribute != nil {
			b.WriteString(fmt.Sprintf("Attribute: %s\n", diag.Attribute))
		}
		b.WriteString("\n")
	}

	return b.String()
}
