// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ReadStateBytesRequest represents a request for the provider to read a data
// source, i.e., update values in state according to the real state of the
// state store. An instance of this request struct is supplied as an argument
// to the state store's Read function.
type ReadStateBytesRequest struct {
	StateID string // The ID of the state to read.
}

type ReadStateResponse struct {
	Bytes       []byte
	Diagnostics diag.Diagnostics
}
