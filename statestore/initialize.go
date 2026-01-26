// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type InitializeClientCapabilities struct {
	ChunkSize int64
}

// TODO: update docs
// -> the ConfigureStateStore RPC itself, similar to provider.Configure
type InitializeRequest struct {
	Config       tfsdk.Config
	ProviderData any
}

// TODO: update docs
type InitializeResponse struct {
	Diagnostics    diag.Diagnostics
	StateStoreData any
}
