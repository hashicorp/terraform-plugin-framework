package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	testErrorDiagnostic1 = diag.NewErrorDiagnostic(
		"Error Diagnostic 1",
		"This is an error.",
	)
	testErrorDiagnostic2 = diag.NewErrorDiagnostic(
		"Error Diagnostic 2",
		"This is an error.",
	)
	testWarningDiagnostic1 = diag.NewWarningDiagnostic(
		"Warning Diagnostic 1",
		"This is a warning.",
	)
	testWarningDiagnostic2 = diag.NewWarningDiagnostic(
		"Warning Diagnostic 2",
		"This is a warning.",
	)
)

type testErrorAttributeValidator struct {
	AttributeValidator
}

func (v testErrorAttributeValidator) Description(ctx context.Context) string {
	return "validation that always returns an error"
}

func (v testErrorAttributeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v testErrorAttributeValidator) Validate(ctx context.Context, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	if len(resp.Diagnostics) == 0 {
		resp.Diagnostics.Append(testErrorDiagnostic1)
	} else {
		resp.Diagnostics.Append(testErrorDiagnostic2)
	}
}

type testWarningAttributeValidator struct {
	AttributeValidator
}

func (v testWarningAttributeValidator) Description(ctx context.Context) string {
	return "validation that always returns a warning"
}

func (v testWarningAttributeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v testWarningAttributeValidator) Validate(ctx context.Context, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	if len(resp.Diagnostics) == 0 {
		resp.Diagnostics.Append(testWarningDiagnostic1)
	} else {
		resp.Diagnostics.Append(testWarningDiagnostic2)
	}
}
