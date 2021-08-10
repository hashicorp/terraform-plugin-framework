package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	testErrorDiagnostic = &tfprotov6.Diagnostic{
		Severity: tfprotov6.DiagnosticSeverityError,
		Summary:  "Error Diagnostic",
		Detail:   "This is an error.",
	}
	testWarningDiagnostic = &tfprotov6.Diagnostic{
		Severity: tfprotov6.DiagnosticSeverityWarning,
		Summary:  "Warning Diagnostic",
		Detail:   "This is a warning.",
	}
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
	resp.Diagnostics = append(resp.Diagnostics, testErrorDiagnostic)
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
	resp.Diagnostics = append(resp.Diagnostics, testWarningDiagnostic)
}
