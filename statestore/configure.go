// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type StateStoreClientCapabilities struct {
	// ChunkSize is the maximum size in bytes for each chunk when reading
	// or writing state. If not set or 0, a default chunk size will be used.
	ChunkSize int64
}

type StateStoreServerCapabilities struct {
	// ChunkSize is the chunk size in bytes that the server will use when
	// reading or writing state. This is typically copied from the client
	// capability or set to a default value.
	ChunkSize int64
}

// ConfigureStateStoreRequest represents a request for the provider to configure an
// state store, i.e., set provider-level data or clients. An instance of this
// request struct is supplied as an argument to the StateStore type Configure
// method.
type ConfigureStateStoreRequest struct {
	// ProviderData is the data set in the
	// [provider.ConfigureResponse.StateStoreData] field. This data is
	// provider-specific and therefore can contain any necessary remote system
	// clients, custom provider data, or anything else pertinent to the
	// functionality of the StateStore.
	//
	// This data is only set after the ConfigureProvider RPC has been called
	// by Terraform.
	ProviderData any

	Config       tfsdk.Config
	Capabilities StateStoreClientCapabilities
}

// ConfigureStateStoreResponse represents a response to a ConfigureStateStoreRequest. An
// instance of this response struct is supplied as an argument to the
// StateStore type Configure method.
type ConfigureStateStoreResponse struct {
	// Diagnostics report errors or warnings related to configuring of the
	// Datasource. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics

	// ServerCapabilities defines optionally supported protocol features for the
	// StateStore, such as chunk size for reading/writing state.
	ServerCapabilities StateStoreServerCapabilities
}
