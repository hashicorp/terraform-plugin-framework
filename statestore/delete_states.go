// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// DeleteStateRequest represents a request for the provider to delete a
// state. An instance of this request struct is supplied as an argument to
// the state's DeleteState function.
type DeleteStateRequest struct {
	TypeName string
	StateId  string
}

// DeleteStateResponse represents a response to a DeleteStateRequest. An
// instance of this response struct is supplied as
// an argument to the state's DeleteStates function, in which the provider
// should set values on the DeleteStateResponse as appropriate.
type DeleteStateResponse struct {
	// Diagnostics report errors or warnings related to deleting the
	// state. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}
