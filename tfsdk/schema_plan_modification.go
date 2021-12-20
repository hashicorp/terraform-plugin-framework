package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-framework/attrpath"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ModifySchemaPlanRequest represents a request for a schema to run all
// attribute plan modification functions.
type ModifySchemaPlanRequest struct {
	// Config is the configuration the user supplied for the resource.
	Config Config

	// State is the current state of the resource.
	State State

	// Plan is the planned new state for the resource.
	Plan Plan

	// ProviderMeta is metadata from the provider_meta block of the module.
	ProviderMeta Config
}

// ModifySchemaPlanResponse represents a response to a ModifySchemaPlanRequest.
type ModifySchemaPlanResponse struct {
	// Plan is the planned new state for the resource.
	Plan Plan

	// RequiresReplace is a list of attrpath.Paths that require the resource to
	// be replaced. They should point to the specific field that changed
	// that requires the resource to be destroyed and recreated.
	RequiresReplace []attrpath.Path

	// Diagnostics report errors or warnings related to running all attribute
	// plan modifiers. Returning an empty slice indicates a successful
	// plan modification with no warnings or errors generated.
	Diagnostics diag.Diagnostics
}
