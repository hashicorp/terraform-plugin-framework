// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"iter"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type WriteStateBytesStream struct {
	Chunks iter.Seq[WriteStateBytesChunk]
}

// WriteStateBytesChunk contains:
//  1. A chunk of state data, received from Terraform core to be persisted.
//  2. Any gRPC-related errors the provider server encountered when
//     receiving data from Terraform core.
//
// If a gRPC error is set, then the chunk should be empty.
type WriteStateBytesChunk struct {
	Meta *WriteStateChunkMeta
	StateByteChunk
	Err error
}

type WriteStateChunkMeta struct {
	TypeName string
	StateId  string
}

type WriteStateBytesResponse struct {
	Diagnostics []*diag.Diagnostic
}
