// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ValidateConfigStateUpgraderRequest represents a request to validate the
// configuration of an state store. An instance of this request struct is
// supplied as an argument to the State Store ValidateConfigStateUpgrader receiver method
// or automatically passed through to each ConfigValidator.
type ValidateConfigStateUpgraderRequest struct {
	// Config is the configuration the user supplied for the state store.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config
}

// ValidateConfigStateUpgraderResponse represents a response to a
// ValidateConfigStateUpgraderRequest. An instance of this response struct is
// supplied as an argument to the State Store ValidateConfigStateUpgrader receiver method
// or automatically passed through to each ConfigValidator.
type ValidateConfigStateUpgraderResponse struct {
	// Diagnostics report errors or warnings related to validating the state store
	// configuration. An empty slice indicates success, with no warnings or
	// errors generated.
	Diagnostics diag.Diagnostics
}
