package reflect

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func toTerraform5ValueErrorDiag(err error, path *tftypes.AttributePath) diag.AttributeErrorDiagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Value Conversion Error",
		"An unexpected error was encountered trying to convert into a Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
	)
}

func toTerraformValueErrorDiag(err error, path *tftypes.AttributePath) diag.AttributeErrorDiagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Value Conversion Error",
		"An unexpected error was encountered trying to convert the Attribute value into a Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
	)
}

func validateValueErrorDiag(err error, path *tftypes.AttributePath) diag.AttributeErrorDiagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Value Conversion Error",
		"An unexpected error was encountered trying to validate the Terraform value type. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
	)
}

func valueFromTerraformErrorDiag(err error, path *tftypes.AttributePath) diag.AttributeErrorDiagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Value Conversion Error",
		"An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
	)
}
