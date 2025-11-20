// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type WriteClientCapabilities struct {
}

// WriteStateBytesRequest represents a request for the provider to read a data
// source, i.e., update values in state according to the real state of the
// state store. An instance of this request struct is supplied as an argument
// to the state store's Write function.
type WriteRequest struct {
	StateId string // The ID of the state to read.
}

type WriteResponse struct {
	Bytes       []byte
	Diagnostics diag.Diagnostics
}
