// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package ephemeral

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// RenewRequest represents a request for the provider to renew an ephemeral
// resource. An instance of this request struct is supplied as an argument to
// the ephemeral resource's Renew function.
type RenewRequest struct {
	// State is the object representing the values of the ephemeral
	// resource following the Open operation.
	State tfsdk.EphemeralState

	// Config is the configuration the user supplied for the ephemeral
	// resource.
	Config tfsdk.Config

	// Private is provider-defined ephemeral resource private state data
	// which was previously provided by the latest Open or Renew operation.
	// Any existing data is copied to RenewResponse.Private to prevent
	// accidental private state data loss.
	//
	// Use the GetKey method to read data. Use the SetKey method on
	// RenewResponse.Private to update or remove a value.
	Private *privatestate.ProviderData
}

// RenewResponse represents a response to a RenewRequest. An
// instance of this response struct is supplied as an argument
// to the ephemeral resource's Renew function, in which the provider
// should set values on the RenewResponse as appropriate.
type RenewResponse struct {
	// RenewAt is an optional date/time field that indicates to Terraform
	// when this ephemeral resource must be renewed at. Terraform will call
	// the (EphemeralResource).Renew method when the current date/time is on
	// or after RenewAt during a Terraform operation.
	//
	// It is recommended to provide small leeway before an ephemeral resource
	// expires, usually no more than a few minutes, to account for clock
	// skew.
	RenewAt time.Time

	// Private is the private state ephemeral resource data following the
	// Renew operation. This field is pre-populated from RenewRequest.Private
	// and can be modified during the ephemeral resource's Renew operation.
	Private *privatestate.ProviderData

	// Diagnostics report errors or warnings related to creating the
	// resource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}
