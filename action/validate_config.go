// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ValidateConfigRequest represents a request to validate the
// configuration of an action. An instance of this request struct is
// supplied as an argument to the Action ValidateConfig receiver method
// or automatically passed through to each ConfigValidator.
type ValidateConfigRequest struct {
	// Config is the configuration the user supplied for the action.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config
}

// ValidateConfigResponse represents a response to a
// ValidateConfigRequest. An instance of this response struct is
// supplied as an argument to the Action ValidateConfig receiver method
// or automatically passed through to each ConfigValidator.
type ValidateConfigResponse struct {
	// Diagnostics report errors or warnings related to validating the action
	// configuration. An empty slice indicates success, with no warnings or
	// errors generated.
	Diagnostics diag.Diagnostics
}
