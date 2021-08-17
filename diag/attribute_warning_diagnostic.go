package diag

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ DiagnosticWithPath = AttributeWarningDiagnostic{}

// AttributeErrorDiagnostic is a generic attribute diagnostic with warning severity.
type AttributeWarningDiagnostic struct {
	WarningDiagnostic

	path *tftypes.AttributePath
}

// Equal returns true if the other diagnostic is wholly equivalent.
func (d AttributeWarningDiagnostic) Equal(other Diagnostic) bool {
	aed, ok := other.(AttributeWarningDiagnostic)

	if !ok {
		return false
	}

	return aed.Summary() == d.Summary() && aed.Detail() == d.Detail() && aed.Path().Equal(d.Path())
}

// Path returns the diagnostic path.
func (d AttributeWarningDiagnostic) Path() *tftypes.AttributePath {
	return d.path
}

// NewAttributeWarningDiagnostic returns a new warning severity diagnostic with the given summary, detail, and path.
func NewAttributeWarningDiagnostic(path *tftypes.AttributePath, summary string, detail string) AttributeWarningDiagnostic {
	return AttributeWarningDiagnostic{
		WarningDiagnostic: WarningDiagnostic{
			detail:  detail,
			summary: summary,
		},
		path: path,
	}
}
