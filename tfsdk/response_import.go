package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ImportResourceStateResponse represents a response to a ImportResourceStateRequest.
// An instance of this response struct is supplied as an argument to the
// Resource's ImportState method, in which the provider should set values on
// the ImportResourceStateResponse as appropriate.
type ImportResourceStateResponse struct {
	// Diagnostics report errors or warnings related to importing the
	// resource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics

	// State is the state of the resource following the import operation.
	// It must contain enough information so Terraform can successfully
	// refresh the resource, e.g. call the Resource Read method.
	State State
}
