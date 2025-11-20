// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// DeleteStatesRequest represents a request for the provider to delete a
// state. An instance of this request struct is supplied as an argument to
// the state's DeleteState function.
type DeleteStatesRequest struct {
	TypeName string
}

// DeleteStatesResponse represents a response to a DeleteStatesRequest. An
// instance of this response struct is supplied as
// an argument to the state's DeleteStates function, in which the provider
// should set values on the DeleteStatesResponse as appropriate.
type DeleteStatesResponse struct {
	// Diagnostics report errors or warnings related to deleting the
	// state. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}
