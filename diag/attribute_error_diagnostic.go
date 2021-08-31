package diag

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ DiagnosticWithPath = AttributeErrorDiagnostic{}

// AttributeErrorDiagnostic is a generic attribute diagnostic with error severity.
type AttributeErrorDiagnostic struct {
	ErrorDiagnostic

	path *tftypes.AttributePath
}

// Equal returns true if the other diagnostic is wholly equivalent.
func (d AttributeErrorDiagnostic) Equal(other Diagnostic) bool {
	aed, ok := other.(AttributeErrorDiagnostic)

	if !ok {
		return false
	}

	if !aed.Path().Equal(d.Path()) {
		return false
	}

	return aed.ErrorDiagnostic.Equal(d.ErrorDiagnostic)
}

// Path returns the diagnostic path.
func (d AttributeErrorDiagnostic) Path() *tftypes.AttributePath {
	return d.path
}

// NewAttributeErrorDiagnostic returns a new error severity diagnostic with the given summary, detail, and path.
func NewAttributeErrorDiagnostic(path *tftypes.AttributePath, summary string, detail string) AttributeErrorDiagnostic {
	return AttributeErrorDiagnostic{
		ErrorDiagnostic: ErrorDiagnostic{
			detail:  detail,
			summary: summary,
		},
		path: path,
	}
}
