// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"iter"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ListRequest represents a request for the provider to list instances of a
// managed resource type that satisfy a user-defined request. An instance of
// this reqeuest struct is passed as an argument to the provider's List
// function implementation.
type ListRequest struct {
	// ListResource is an instance of the provider's ListResource
	// implementation for a specific managed resource type.
	ListResource list.ListResource

	// Config is the configuration the user supplied for listing resource
	// instances.
	Config tfsdk.Config

	// IncludeResource indicates whether the provider should populate the
	// Resource field in the ListResult struct.
	IncludeResource bool
}

// ListResultsStream represents a streaming response to a ListRequest.  An
// instance of this struct is supplied as an argument to the provider's List
// function. The provider should set a Results iterator function that yields
// zero or more results of type ListResult.
//
// For convenience, a provider implementation may choose to convert a slice of
// results into an iterator using [slices.Values].
//
// [slices.Values]: https://pkg.go.dev/slices#Values
type ListResourceStream struct {
	// Results is a function that emits ListResult values via its yield
	// function argument.
	Results iter.Seq[ListResult]
}

// ListResult represents a listed managed resource instance.
type ListResult struct {
	// Identity is the identity of the managed resource instance.  A nil value
	// will raise will raise a diagnostic.
	Identity *tfsdk.ResourceIdentity

	// Resource is the provider's representation of the attributes of the
	// listed managed resource instance.
	//
	// If ListRequest.IncludeResource is true, a nil value will raise
	// a warning diagnostic.
	Resource *tfsdk.Resource

	// DisplayName is a provider-defined human-readable description of the
	// listed managed resource instance, intended for CLI and browser UIs.
	DisplayName string

	// Diagnostics report errors or warnings related to the listed managed
	// resource instance. An empty slice indicates a successful operation with
	// no warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// ListResource implements the framework server ListResource RPC.
func (s *Server) ListResource(ctx context.Context, fwReq *ListRequest, fwStream *ListResourceStream) {
	listResource := fwReq.ListResource

	req := list.ListRequest{
		Config:          fwReq.Config,
		IncludeResource: fwReq.IncludeResource,
	}

	stream := &list.ListResultsStream{}

	logging.FrameworkTrace(ctx, "Calling provider defined ListResource")
	listResource.List(ctx, req, stream)
	logging.FrameworkTrace(ctx, "Called provider defined ListResource")

	if stream.Results == nil {
		// If the provider returned a nil results stream, we treat it as an empty stream.
		stream.Results = func(func(list.ListResult) bool) {}
	}

	fwStream.Results = listResourceEventStreamAdapter(stream.Results)
}

func listResourceEventStreamAdapter(stream iter.Seq[list.ListResult]) iter.Seq[ListResult] {
	// TODO: is this any more efficient than a for-range?
	return func(yieldFw func(ListResult) bool) {
		yield := func(event list.ListResult) bool {
			return yieldFw(ListResult(event))
		}
		stream(yield)
	}
}
