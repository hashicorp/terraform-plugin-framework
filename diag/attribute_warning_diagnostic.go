package diag

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// NewAttributeWarningDiagnostic returns a new warning severity diagnostic with the given summary, detail, and path.
func NewAttributeWarningDiagnostic(path *tftypes.AttributePath, summary string, detail string) DiagnosticWithPath {
	return withPath{
		Diagnostic: NewWarningDiagnostic(summary, detail),
		path:       path,
	}
}
