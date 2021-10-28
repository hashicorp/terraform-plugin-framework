package diag

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ Diagnostic = &GenericDiagnostic{}

// GenericDiagnostic is a built-in diagnostic type that can be extended.
type GenericDiagnostic struct {
	severity Severity
	detail   string
	summary  string
	path     *tftypes.AttributePath
}

// Detail returns the diagnostic detail.
func (d GenericDiagnostic) Detail() string {
	return d.detail
}

// Equal returns true if the other diagnostic is wholly equivalent.
func (d *GenericDiagnostic) Equal(other Diagnostic) bool {
	if d == nil && other == nil {
		return true
	}

	gd, ok := other.(*GenericDiagnostic)

	if !ok {
		return false
	}

	if gd.Severity() != d.Severity() {
		return false
	}

	if gd.Summary() != d.Summary() {
		return false
	}

	if gd.Detail() != d.Detail() {
		return false
	}

	return gd.Path().Equal(d.Path())
}

// Path returns the diagnostic path.
func (d GenericDiagnostic) Path() *tftypes.AttributePath {
	return d.path
}

// SetPath sets the diagnostic path.
func (d *GenericDiagnostic) SetPath(path *tftypes.AttributePath) {
	d.path = path
}

// Severity returns the diagnostic severity.
func (d GenericDiagnostic) Severity() Severity {
	return d.severity
}

// Summary returns the diagnostic summary.
func (d GenericDiagnostic) Summary() string {
	return d.summary
}

// NewAttributeErrorDiagnostic returns a new error severity diagnostic with the given summary, detail, and path.
func NewAttributeErrorDiagnostic(path *tftypes.AttributePath, summary string, detail string) *GenericDiagnostic {
	return &GenericDiagnostic{
		detail:   detail,
		path:     path,
		severity: SeverityError,
		summary:  summary,
	}
}

// NewAttributeWarningDiagnostic returns a new warning severity diagnostic with the given summary, detail, and path.
func NewAttributeWarningDiagnostic(path *tftypes.AttributePath, summary string, detail string) *GenericDiagnostic {
	return &GenericDiagnostic{
		detail:   detail,
		path:     path,
		severity: SeverityWarning,
		summary:  summary,
	}
}

// NewErrorDiagnostic returns a new error severity diagnostic with the given summary and detail.
func NewErrorDiagnostic(summary string, detail string) *GenericDiagnostic {
	return &GenericDiagnostic{
		detail:   detail,
		severity: SeverityError,
		summary:  summary,
	}
}

// NewWarningDiagnostic returns a new warning severity diagnostic with the given summary and detail.
func NewWarningDiagnostic(summary string, detail string) *GenericDiagnostic {
	return &GenericDiagnostic{
		detail:   detail,
		severity: SeverityWarning,
		summary:  summary,
	}
}
