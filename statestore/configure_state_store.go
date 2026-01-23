// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type ConfigureStateStoreClientCapabilities struct {
	ChunkSize int64
}

// TODO: update docs
// -> the RPC itself, similar to provider.Configure
type ConfigureStateStoreRequest struct {
	Config       tfsdk.Config
	ProviderData any
}

// TODO: update docs
type ConfigureStateStoreResponse struct {
	Diagnostics    diag.Diagnostics
	StateStoreData any
}
