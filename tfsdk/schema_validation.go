package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ValidateSchemaRequest repesents a request for validating a Schema.
type ValidateSchemaRequest struct {
	// Config contains the entire configuration of the data source, provider, or resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config Config
}

// ValidateSchemaResponse represents a response to a
// ValidateSchemaRequest.
type ValidateSchemaResponse struct {
	// Diagnostics report errors or warnings related to validating the schema.
	// An empty slice indicates success, with no warnings or errors generated.
	Diagnostics diag.Diagnostics
}
