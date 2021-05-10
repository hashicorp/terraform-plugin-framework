package tf

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ConfigureProviderResponse struct {
	Diagnostics []*tfprotov6.Diagnostic
}

type CreateResourceResponse struct {
	NewState State

	Diagnostics []*tfprotov6.Diagnostic
}

func (r *CreateResourceResponse) AddWarning(summary, detail string) {
	r.Diagnostics = append(r.Diagnostics, &tfprotov6.Diagnostic{
		Summary:  summary,
		Detail:   detail,
		Severity: tfprotov6.DiagnosticSeverityWarning,
	})
}

func (r *CreateResourceResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics = append(r.Diagnostics, &tfprotov6.Diagnostic{
		Attribute: attributePath,
		Summary:   summary,
		Detail:    detail,
		Severity:  tfprotov6.DiagnosticSeverityWarning,
	})
}

func (r *CreateResourceResponse) AddError(summary, detail string) {
	r.Diagnostics = append(r.Diagnostics, &tfprotov6.Diagnostic{
		Summary:  summary,
		Detail:   detail,
		Severity: tfprotov6.DiagnosticSeverityError,
	})
}

func (r *CreateResourceResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics = append(r.Diagnostics, &tfprotov6.Diagnostic{
		Attribute: attributePath,
		Summary:   summary,
		Detail:    detail,
		Severity:  tfprotov6.DiagnosticSeverityError,
	})
}

// func (r *CreateResourceResponse) SetNewState(attributePath tftypes.AttributePath, val interface{}) error {
// 	return r.NewState.Set(attributePath, val)
// }

type ReadResourceResponse struct {
	Diagnostics []*tfprotov6.Diagnostic
}

type DeleteResourceResponse struct {
	Diagnostics []*tfprotov6.Diagnostic
}

type UpdateResourceResponse struct {
	NewState State

	Diagnostics []*tfprotov6.Diagnostic
}
