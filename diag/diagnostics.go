package diag

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Diagnostics represents a collection of diagnostics.
//
// While this collection is ordered, the order is not guaranteed as reliable
// or consistent.
type Diagnostics []Diagnostic

// AddAttributeError adds a generic attribute error diagnostic to the collection.
func (diags *Diagnostics) AddAttributeError(path *tftypes.AttributePath, summary string, detail string) {
	diags.Append(NewAttributeErrorDiagnostic(path, summary, detail))
}

// AddAttributeWarning adds a generic attribute warning diagnostic to the collection.
func (diags *Diagnostics) AddAttributeWarning(path *tftypes.AttributePath, summary string, detail string) {
	diags.Append(NewAttributeWarningDiagnostic(path, summary, detail))
}

// AddError adds a generic error diagnostic to the collection.
func (diags *Diagnostics) AddError(summary string, detail string) {
	diags.Append(NewErrorDiagnostic(summary, detail))
}

// AddWarning adds a generic warning diagnostic to the collection.
func (diags *Diagnostics) AddWarning(summary string, detail string) {
	diags.Append(NewWarningDiagnostic(summary, detail))
}

// Append adds non-empty and non-duplicate diagnostics to the collection.
func (diags *Diagnostics) Append(in ...Diagnostic) {
	for _, diag := range in {
		if diag == nil {
			continue
		}

		if diags.Contains(diag) {
			continue
		}

		if diags == nil {
			*diags = Diagnostics{diag}
		} else {
			*diags = append(*diags, diag)
		}
	}
}

// Contains returns true if the collection contains an equal Diagnostic.
func (diags Diagnostics) Contains(in Diagnostic) bool {
	for _, diag := range diags {
		if diag.Equal(in) {
			return true
		}
	}

	return false
}

// HasError returns true if the collection has an error severity Diagnostic.
func (diags Diagnostics) HasError() bool {
	for _, diag := range diags {
		if diag.Severity() == SeverityError {
			return true
		}
	}

	return false
}

// ToTfprotov6Diagnostics converts the diagnostics into the tfprotov6 collection type.
//
// Usage of this method outside the framework is not supported nor considered
// for backwards compatibility promises.
func (diags Diagnostics) ToTfprotov6Diagnostics() []*tfprotov6.Diagnostic {
	var results []*tfprotov6.Diagnostic

	for _, diag := range diags {
		tfprotov6Diagnostic := &tfprotov6.Diagnostic{
			Detail:   diag.Detail(),
			Severity: diag.Severity().ToTfprotov6DiagnosticSeverity(),
			Summary:  diag.Summary(),
		}

		if diagWithPath, ok := diag.(DiagnosticWithPath); ok {
			tfprotov6Diagnostic.Attribute = diagWithPath.Path()
		}

		results = append(results, tfprotov6Diagnostic)
	}

	return results
}
