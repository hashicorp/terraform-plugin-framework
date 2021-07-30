package reflect

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

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

func toTerraform5ValueErrorDiag(err error, path *tftypes.AttributePath) *tfprotov6.Diagnostic {
	return &tfprotov6.Diagnostic{
		Severity:  tfprotov6.DiagnosticSeverityError,
		Summary:   "Value Conversion Error",
		Detail:    "An unexpected error was encountered trying to convert into a Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		Attribute: path,
	}
}

func toTerraformValueErrorDiag(err error, path *tftypes.AttributePath) *tfprotov6.Diagnostic {
	return &tfprotov6.Diagnostic{
		Severity:  tfprotov6.DiagnosticSeverityError,
		Summary:   "Value Conversion Error",
		Detail:    "An unexpected error was encountered trying to convert the Attribute value into a Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		Attribute: path,
	}
}

func validateValueErrorDiag(err error, path *tftypes.AttributePath) *tfprotov6.Diagnostic {
	return &tfprotov6.Diagnostic{
		Severity:  tfprotov6.DiagnosticSeverityError,
		Summary:   "Value Conversion Error",
		Detail:    "An unexpected error was encountered trying to validate the Terraform value type. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		Attribute: path,
	}
}

func valueFromTerraformErrorDiag(err error, path *tftypes.AttributePath) *tfprotov6.Diagnostic {
	return &tfprotov6.Diagnostic{
		Severity:  tfprotov6.DiagnosticSeverityError,
		Summary:   "Value Conversion Error",
		Detail:    "An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		Attribute: path,
	}
}
