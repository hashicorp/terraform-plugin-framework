// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package ephemeral

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// OpenRequest represents a request for the provider to open an ephemeral
// resource. An instance of this request struct is supplied as an argument to
// the ephemeral resource's Open function.
type OpenRequest struct {
	// Config is the configuration the user supplied for the ephemeral
	// resource.
	Config tfsdk.Config
}

// OpenResponse represents a response to a OpenRequest. An
// instance of this response struct is supplied as an argument
// to the ephemeral resource's Open function, in which the provider
// should set values on the OpenResponse as appropriate.
type OpenResponse struct {
	// State is the object representing the values of the ephemeral
	// resource following the Open operation. This field is pre-populated
	// from OpenRequest.Config and should be set during the resource's Open
	// operation.
	State tfsdk.EphemeralState

	// Private is the private state ephemeral resource data following the
	// Open operation. This field is not pre-populated as there is no
	// pre-existing private state data during the ephemeral resource's
	// Open operation.
	Private *privatestate.ProviderData

	// RenewAt is an optional date/time field that indicates to Terraform
	// when this ephemeral resource must be renewed at. Terraform will call
	// the (EphemeralResource).Renew method when the current date/time is on
	// or after RenewAt during a Terraform operation.
	//
	// It is recommended to provide small leeway before an ephemeral resource
	// expires, usually no more than a few minutes, to account for clock
	// skew.
	RenewAt time.Time

	// Diagnostics report errors or warnings related to creating the
	// resource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}
