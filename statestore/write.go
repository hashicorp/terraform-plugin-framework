// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// WriteClientCapabilities allows Terraform to publish information
// regarding optionally supported protocol features for the WriteStateStore RPC,
// such as forward-compatible Terraform behavior changes.
type WriteClientCapabilities struct {
}

// WriteRequest represents a request for the provider to write a data
// source, i.e., update values in state according to the real state of the
// state store. An instance of this request struct is supplied as an argument
// to the state store's Write function.
type WriteRequest struct {
	// Config is the configuration the user supplied for the state store.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config

	// ProviderMeta is metadata from the provider_meta block of the module.
	ProviderMeta tfsdk.Config

	// ClientCapabilities defines optionally supported protocol features for the
	// WriteStateStore RPC, such as forward-compatible Terraform behavior changes.
	ClientCapabilities WriteClientCapabilities
}

// WriteResponse represents a response to a WriteRequest. An
// instance of this response struct is supplied as an argument to the data
// source's Write function, in which the provider should set values on the
// WriteResponse as appropriate.
type WriteResponse struct {
	// Diagnostics report errors or warnings related to writeing the data
	// source. An empty slice indicates a successful operation with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}
