// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
)

// DeleteStatesRequest represents a request for the provider to delete a
// state. An instance of this request struct is supplied as an argument to
// the state's DeleteStates function.
type DeleteStatesRequest struct {
	// Private is provider-defined state private state data which was previously
	// stored with the state state.
	//
	// Use the GetKey method to read data.
	Private *privatestate.ProviderData
}

// DeleteStatesResponse represents a response to a DeleteStatesRequest. An
// instance of this response struct is supplied as
// an argument to the state's DeleteStates function, in which the provider
// should set values on the DeleteStatesResponse as appropriate.
type DeleteStatesResponse struct {
	// Private is the private state data following the DeleteStates
	// operation. This field is pre-populated from DeleteStatesRequest.Private and
	// can be modified during the state's DeleteStates operation in cases where
	// an error diagnostic is being returned. Otherwise if no error diagnostic
	// is being returned, indicating that the state was successfully deleted,
	// this data will be automatically cleared to prevent Terraform errors.
	Private *privatestate.ProviderData

	// Diagnostics report errors or warnings related to deleting the
	// state. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}
