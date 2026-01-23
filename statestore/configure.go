// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// TODO: update docs
// -> passing global state store configured data (returned from ConfigureStateStore, stored on provider server) to
// an implementation that is about to run a state store RPC (i.e. ReadStateBytes)
type ConfigureRequest struct {
	StateStoreData any
}

// TODO: update docs
type ConfigureResponse struct {
	Diagnostics diag.Diagnostics
}
