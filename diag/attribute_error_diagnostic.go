package diag

import "github.com/hashicorp/terraform-plugin-framework/attrpath"

// NewAttributeErrorDiagnostic returns a new error severity diagnostic with the given summary, detail, and path.
func NewAttributeErrorDiagnostic(path attrpath.Path, summary string, detail string) DiagnosticWithPath {
	return withPath{
		Diagnostic: NewErrorDiagnostic(summary, detail),
		path:       path,
	}
}
