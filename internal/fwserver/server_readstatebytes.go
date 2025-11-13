// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"iter"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ReadStateBytesRequest is the framework server request for the StateBytesResource RPC.
type ReadStateBytesRequest struct {
	StateStore statestore.StateStore
	StateId    string
}

// ReadStateBytesStream represents a streaming response to a [ReadStateBytesRequest].  An
// instance of this struct is supplied as an argument to the provider's StateBytes
// function. The provider should set a Chunks iterator function that pushes
// zero or more results of type [StateBytesResult].
//
// For convenience, a provider implementation may choose to convert a slice of
// results into an iterator using [slices.Values].
type ReadStateBytesStream struct {
	// Chunks is a function that emits [StateBytesResult] values via its push
	// function argument.
	Chunks iter.Seq[tfprotov6.ReadStateByteChunk]
}

func StateBytesResultError(summary string, detail string) StateBytesResult {
	return StateBytesResult{
		Diagnostics: diag.Diagnostics{
			diag.NewErrorDiagnostic(summary, detail),
		},
	}
}

// StateBytesResult represents a state store.
type StateBytesResult struct {
	// Diagnostics report errors or warnings related to the statebytes managed
	// resource instance. An empty slice indicates a successful operation with
	// no warnings or errors generated.
	Diagnostics diag.Diagnostics
}

var NoStateBytesChunks = func(func(StateBytesResult) bool) {}

// StateBytesResource implements the framework server StateBytesResource RPC.
func (s *Server) StateBytesResource(ctx context.Context, fwReq *ReadStateBytesRequest, fwStream *ReadStateBytesStream) {
	diagsStream := &statestore.ReadStateBytesStream{}

	req := statestore.ReadStateBytesRequest{
		TypeName: fwReq.TypeName,
		StateId:  fwReq.StateId,
	}

	stream := &statestore.ReadStateBytesStream{}

	// If the provider returned a nil results stream, we return an empty stream.
	if diagsStream.Chunks == nil {
		diagsStream.Chunks = statestore.NoStateBytesChunks
	}

	if stream.Chunks == nil {
		stream.Chunks = statestore.NoStateBytesChunks
	}

	fwStream.Chunks = processStateBytesChunks(req, stream.Chunks, diagsStream.Chunks)
}

func processStateBytesChunks(req statestore.ReadStateBytesRequest, streams ...iter.Seq[statestore.StateBytesResult]) iter.Seq[StateBytesResult] {
	return func(push func(StateBytesResult) bool) {
		for _, stream := range streams {
			for result := range stream {
				if !push(processStateBytesResult(req, result)) {
					return
				}
			}
		}
	}
}

// processStateBytesResult validates the content of a statestore.StateBytesResult and returns a
// StateBytesResult
func processStateBytesResult(req statestore.ReadStateBytesRequest, result statestore.StateByteChunk) StateBytesResult {
	if result.Diagnostics.HasError() {
		return StateBytesResult(result)
	}

	// Allow any non-error diags to pass through
	if len(result.Diagnostics) > 0 {
		return StateBytesResult(result)
	}

	return StateBytesResult(result)
}
