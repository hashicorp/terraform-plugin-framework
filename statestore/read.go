// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"iter"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ReadClientCapabilities allows Terraform to publish information
// regarding optionally supported protocol features for the ReadStateStore RPC,
// such as forward-compatible Terraform behavior changes.
type ReadClientCapabilities struct {
}

// ReadStateBytesRequest represents a request for the provider to read a data
// source, i.e., update values in state according to the real state of the
// state store. An instance of this request struct is supplied as an argument
// to the state store's Read function.
type ReadStateBytesRequest struct {
	TypeName string
	StateId  string
}

// ReadStateBytesStream represents a response to a ReadStateBytesRequest. An
// instance of this response struct is supplied as an argument to the data
// source's Read function, in which the provider should set values on the
// ReadStateBytesStream as appropriate.
type ReadStateBytesStream struct {
	Chunks iter.Seq[ReadStateByteChunk]
}

type ReadStateByteChunk struct {
	StateByteChunk
	Diagnostics []*diag.Diagnostic
}
type StateByteChunk struct {
	Bytes       []byte
	TotalLength int64
	Range       StateByteRange
}

type StateByteRange struct {
	Start, End int64
}
