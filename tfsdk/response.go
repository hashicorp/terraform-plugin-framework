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
