package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ValidateDataSourceConfigResponse represents a response to a
// ValidateDataSourceConfigRequest. An instance of this response struct is
// supplied as an argument to the DataSource ValidateConfig receiver method
// or automatically passed through to each ConfigValidator.
type ValidateDataSourceConfigResponse struct {
	// Diagnostics report errors or warnings related to validating the data
	// source configuration. An empty slice indicates success, with no warnings
	// or errors generated.
	Diagnostics diag.Diagnostics
}

// ValidateResourceConfigResponse represents a response to a
// ValidateResourceConfigRequest. An instance of this response struct is
// supplied as an argument to the Resource ValidateConfig receiver method
// or automatically passed through to each ConfigValidator.
type ValidateResourceConfigResponse struct {
	// Diagnostics report errors or warnings related to validating the resource
	// configuration. An empty slice indicates success, with no warnings or
	// errors generated.
	Diagnostics diag.Diagnostics
}

// ValidateProviderConfigResponse represents a response to a
// ValidateProviderConfigRequest. An instance of this response struct is
// supplied as an argument to the Provider ValidateConfig receiver method
// or automatically passed through to each ConfigValidator.
type ValidateProviderConfigResponse struct {
	// Diagnostics report errors or warnings related to validating the provider
	// configuration. An empty slice indicates success, with no warnings or
	// errors generated.
	Diagnostics diag.Diagnostics
}
