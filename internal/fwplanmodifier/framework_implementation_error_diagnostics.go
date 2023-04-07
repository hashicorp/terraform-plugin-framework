package fwplanmodifier

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// FrameworkImplementationErrorDiag returns an error diagnostic intended for
// when a framework-defined schema plan modifier reached an unexpected
// implementation issue.
func FrameworkImplementationErrorDiag(p path.Path, details string) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		p,
		"Framework Plan Modifier Implementation Error",
		"A framework-defined plan modifier encountered an unexpected implementation issue which could cause unexpected behavior or panics. "+
			"This is always an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("Path: %s\n", p)+
			"Details: "+details,
	)
}

// PlanValueTypeAssertionDiag returns an error diagnostic intended for when a
// schema plan modifier with a shared implementation did not return the expected
// type in the response for the typed response.
func PlanValueTypeAssertionDiag(p path.Path, requestType attr.Value, responseType attr.Value) diag.Diagnostic {
	return FrameworkImplementationErrorDiag(
		p,
		"The shared implementation responded with an unexpected type.\n"+
			fmt.Sprintf("Expected Type: %T\n", requestType)+
			fmt.Sprintf("Response Type: %T", responseType),
	)
}
