package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ValidateSchemaRequest repesents a request for validating a Schema.
type ValidateSchemaRequest struct {
	// Config contains the entire configuration of the data source, provider, or resource.
	Config tfsdk.Config
}

// ValidateSchemaResponse represents a response to a
// ValidateSchemaRequest.
type ValidateSchemaResponse struct {
	// Diagnostics report errors or warnings related to validating the data
	// source configuration. An empty slice indicates success, with no warnings
	// or errors generated.
	Diagnostics []*tfprotov6.Diagnostic
}
