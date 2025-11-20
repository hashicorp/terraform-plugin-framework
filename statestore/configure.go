// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type StateStoreClientCapabilities struct {
	DeferralAllowed bool
}

// ConfigureStateStoreRequest represents a request for the provider to configure an
// state store, i.e., set provider-level data or clients. An instance of this
// request struct is supplied as an argument to the StateStore type Configure
// method.
type ConfigureStateStoreRequest struct {
	// ProviderData is the data set in the
	// [provider.ConfigureStateStoreResponse.StateStoreData] field. This data is
	// provider-specific and therefore can contain any necessary remote system
	// clients, custom provider data, or anything else pertinent to the
	// functionality of the StateStore.
	//
	// This data is only set after the ConfigureProvider RPC has been called
	// by Terraform.
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
}
