package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ConfigureProviderResponse represents a response to a
// ConfigureProviderRequest. An instance of this response struct is supplied as
// an argument to the provider's Configure function, in which the provider
// should set values on the ConfigureProviderResponse as appropriate.
type ConfigureProviderResponse struct {
	// Diagnostics report errors or warnings related to configuring the
	// provider. An empty slice indicates success, with no warnings or
	// errors generated.
	Diagnostics diag.Diagnostics
}

// AddWarning appends a warning diagnostic to the response. If the warning
// concerns a particular attribute, AddAttributeWarning should be used instead.
func (r *ConfigureProviderResponse) AddWarning(summary, detail string) {
	r.Diagnostics.AddWarning(summary, detail)
}

// AddAttributeWarning appends a warning diagnostic to the response and labels
// it with a specific attribute.
func (r *ConfigureProviderResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeWarning(attributePath, summary, detail)
}

// AddError appends an error diagnostic to the response. If the error concerns a
// particular attribute, AddAttributeError should be used instead.
func (r *ConfigureProviderResponse) AddError(summary, detail string) {
	r.Diagnostics.AddError(summary, detail)
}

// AddAttributeError appends an error diagnostic to the response and labels it
// with a specific attribute.
func (r *ConfigureProviderResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeError(attributePath, summary, detail)
}

// CreateResourceResponse represents a response to a CreateResourceRequest. An
// instance of this response struct is supplied as
// an argument to the resource's Create function, in which the provider
// should set values on the CreateResourceResponse as appropriate.
type CreateResourceResponse struct {
	// State is the state of the resource following the Create operation.
	// This field is pre-populated from CreateResourceRequest.Plan and
	// should be set during the resource's Create operation.
	State State

	// Diagnostics report errors or warnings related to creating the
	// resource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// AddWarning appends a warning diagnostic to the response. If the warning
// concerns a particular attribute, AddAttributeWarning should be used instead.
func (r *CreateResourceResponse) AddWarning(summary, detail string) {
	r.Diagnostics.AddWarning(summary, detail)
}

// AddAttributeWarning appends a warning diagnostic to the response and labels
// it with a specific attribute.
func (r *CreateResourceResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeWarning(attributePath, summary, detail)
}

// AddError appends an error diagnostic to the response. If the error concerns a
// particular attribute, AddAttributeError should be used instead.
func (r *CreateResourceResponse) AddError(summary, detail string) {
	r.Diagnostics.AddError(summary, detail)
}

// AddAttributeError appends an error diagnostic to the response and labels it
// with a specific attribute.
func (r *CreateResourceResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeError(attributePath, summary, detail)
}

// ReadResourceResponse represents a response to a ReadResourceRequest. An
// instance of this response struct is supplied as
// an argument to the resource's Read function, in which the provider
// should set values on the ReadResourceResponse as appropriate.
type ReadResourceResponse struct {
	// State is the state of the resource following the Read operation.
	// This field is pre-populated from ReadResourceRequest.State and
	// should be set during the resource's Read operation.
	State State

	// Diagnostics report errors or warnings related to reading the
	// resource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// AddWarning appends a warning diagnostic to the response. If the warning
// concerns a particular attribute, AddAttributeWarning should be used instead.
func (r *ReadResourceResponse) AddWarning(summary, detail string) {
	r.Diagnostics.AddWarning(summary, detail)
}

// AddAttributeWarning appends a warning diagnostic to the response and labels
// it with a specific attribute.
func (r *ReadResourceResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeWarning(attributePath, summary, detail)
}

// AddError appends an error diagnostic to the response. If the error concerns a
// particular attribute, AddAttributeError should be used instead.
func (r *ReadResourceResponse) AddError(summary, detail string) {
	r.Diagnostics.AddError(summary, detail)
}

// AddAttributeError appends an error diagnostic to the response and labels it
// with a specific attribute.
func (r *ReadResourceResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeError(attributePath, summary, detail)
}

// UpdateResourceResponse represents a response to an UpdateResourceRequest. An
// instance of this response struct is supplied as
// an argument to the resource's Update function, in which the provider
// should set values on the UpdateResourceResponse as appropriate.
type UpdateResourceResponse struct {
	// State is the state of the resource following the Update operation.
	// This field is pre-populated from UpdateResourceRequest.Plan and
	// should be set during the resource's Update operation.
	State State

	// Diagnostics report errors or warnings related to updating the
	// resource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// AddWarning appends a warning diagnostic to the response. If the warning
// concerns a particular attribute, AddAttributeWarning should be used instead.
func (r *UpdateResourceResponse) AddWarning(summary, detail string) {
	r.Diagnostics.AddWarning(summary, detail)
}

// AddAttributeWarning appends a warning diagnostic to the response and labels
// it with a specific attribute.
func (r *UpdateResourceResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeWarning(attributePath, summary, detail)
}

// AddError appends an error diagnostic to the response. If the error concerns a
// particular attribute, AddAttributeError should be used instead.
func (r *UpdateResourceResponse) AddError(summary, detail string) {
	r.Diagnostics.AddError(summary, detail)
}

// AddAttributeError appends an error diagnostic to the response and labels it
// with a specific attribute.
func (r *UpdateResourceResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeError(attributePath, summary, detail)
}

// DeleteResourceResponse represents a response to a DeleteResourceRequest. An
// instance of this response struct is supplied as
// an argument to the resource's Delete function, in which the provider
// should set values on the DeleteResourceResponse as appropriate.
type DeleteResourceResponse struct {
	// State is the state of the resource following the Delete operation.
	// This field is pre-populated from UpdateResourceRequest.Plan and
	// should be set during the resource's Update operation.
	State State

	// Diagnostics report errors or warnings related to deleting the
	// resource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// AddWarning appends a warning diagnostic to the response. If the warning
// concerns a particular attribute, AddAttributeWarning should be used instead.
func (r *DeleteResourceResponse) AddWarning(summary, detail string) {
	r.Diagnostics.AddWarning(summary, detail)
}

// AddAttributeWarning appends a warning diagnostic to the response and labels
// it with a specific attribute.
func (r *DeleteResourceResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeWarning(attributePath, summary, detail)
}

// AddError appends an error diagnostic to the response. If the error concerns a
// particular attribute, AddAttributeError should be used instead.
func (r *DeleteResourceResponse) AddError(summary, detail string) {
	r.Diagnostics.AddError(summary, detail)
}

// AddAttributeError appends an error diagnostic to the response and labels it
// with a specific attribute.
func (r *DeleteResourceResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeError(attributePath, summary, detail)
}

// ModifyResourcePlanResponse represents a response to a
// ModifyResourcePlanRequest. An instance of this response struct is supplied
// as an argument to the resource's ModifyPlan function, in which the provider
// should modify the Plan and populate the RequiresReplace field as appropriate.
type ModifyResourcePlanResponse struct {
	// Plan is the planned new state for the resource.
	Plan Plan

	// RequiresReplace is a list of tftypes.AttributePaths that require the
	// resource to be replaced. They should point to the specific field
	// that changed that requires the resource to be destroyed and
	// recreated.
	RequiresReplace []*tftypes.AttributePath

	// Diagnostics report errors or warnings related to determining the
	// planned state of the requested resource. Returning an empty slice
	// indicates a successful plan modification with no warnings or errors
	// generated.
	Diagnostics diag.Diagnostics
}

// AddWarning appends a warning diagnostic to the response. If the warning
// concerns a particular attribute, AddAttributeWarning should be used instead.
func (r *ModifyResourcePlanResponse) AddWarning(summary, detail string) {
	r.Diagnostics.AddWarning(summary, detail)
}

// AddAttributeWarning appends a warning diagnostic to the response and labels
// it with a specific attribute.
func (r *ModifyResourcePlanResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeWarning(attributePath, summary, detail)
}

// AddError appends an error diagnostic to the response. If the error concerns a
// particular attribute, AddAttributeError should be used instead.
func (r *ModifyResourcePlanResponse) AddError(summary, detail string) {
	r.Diagnostics.AddError(summary, detail)
}

// AddAttributeError appends an error diagnostic to the response and labels it
// with a specific attribute.
func (r *ModifyResourcePlanResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeError(attributePath, summary, detail)
}

// ReadDataSourceResponse represents a response to a ReadDataSourceRequest. An
// instance of this response struct is supplied as an argument to the data
// source's Read function, in which the provider should set values on the
// ReadDataSourceResponse as appropriate.
type ReadDataSourceResponse struct {
	// State is the state of the data source following the Read operation.
	// This field should be set during the resource's Read operation.
	State State

	// Diagnostics report errors or warnings related to reading the data
	// source. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// AddWarning appends a warning diagnostic to the response. If the warning
// concerns a particular attribute, AddAttributeWarning should be used instead.
func (r *ReadDataSourceResponse) AddWarning(summary, detail string) {
	r.Diagnostics.AddWarning(summary, detail)
}

// AddAttributeWarning appends a warning diagnostic to the response and labels
// it with a specific attribute.
func (r *ReadDataSourceResponse) AddAttributeWarning(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeWarning(attributePath, summary, detail)
}

// AddError appends an error diagnostic to the response. If the error concerns a
// particular attribute, AddAttributeError should be used instead.
func (r *ReadDataSourceResponse) AddError(summary, detail string) {
	r.Diagnostics.AddError(summary, detail)
}

// AddAttributeError appends an error diagnostic to the response and labels it
// with a specific attribute.
func (r *ReadDataSourceResponse) AddAttributeError(attributePath *tftypes.AttributePath, summary, detail string) {
	r.Diagnostics.AddAttributeError(attributePath, summary, detail)
}
