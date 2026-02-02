// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ReadStateBytesRequest represents a request containing the values the user
// specified for the state_store configuration block, along with the data configured
// for the provider itself, populated by the [provider.ConfigureResponse.StateStoreData] field.
//
// An instance of this request struct is supplied as an argument to the state store's
// Read method.
type ReadStateBytesRequest struct {
	StateID string // The ID of the state to read.

	// ProviderData is the data set in the [provider.ConfigureResponse.StateStoreData]
	// field. This data is provider-specific and therefore can contain any necessary remote system
	// clients, custom provider data, or anything else pertinent to the functionality of the StateStore.
	ProviderData any
}

// ReadStateBytesResponse represents a response to an ReadStateBytesRequest. An instance of this response
// struct is supplied as an argument to the state store's Read method, in which the provider
// should set values on the ReadStateBytesResponse as appropriate.
type ReadStateBytesResponse struct {
	// Diagnostics report errors or warnings related to initializing the
	// state store. An empty slice indicates success, with no warnings or
	// errors generated.
	Diagnostics diag.Diagnostics

	StateBytes []byte
}
