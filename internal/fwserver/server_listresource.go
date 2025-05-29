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

type ListResourceRequest struct {
	// ListResource is ...
	ListResource list.ListResource

	Config tfsdk.Config

	IncludeResourceObject bool
}

type ListResourceStream struct {
	Results     iter.Seq[ListResourceEvent]
	Diagnostics diag.Diagnostics
}

type ListResourceEvent struct {
	// Identity is the identity of the managed resource instance.
	// A nil value will raise a diagnostic.
	Identity *tfsdk.ResourceIdentity

	// ResourceObject is the provider's representation of the attributes of the
	// listed managed resource instance.
	// If ListResourceRequest.IncludeResourceObject is true, a nil value will raise
	// a warning diagnostic.
	ResourceObject *tfsdk.ResourceObject

	// DisplayName is a provider-defined human-readable description of the
	// listed managed resource instance, intended for CLI and browser UIs.
	DisplayName string

	// Diagnostics report errors or warnings related to the listed managed
	// resource instance. An empty slice indicates a successful operation with
	// no warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// ListResource implements the framework server ListResource RPC.
func (s *Server) ListResource(ctx context.Context, fwReq *ListResourceRequest, fwStream *ListResourceStream) {
	listResource := fwReq.ListResource

	req := list.ListResourceRequest{
		Config:                fwReq.Config,
		IncludeResourceObject: fwReq.IncludeResourceObject,
	}

	stream := &list.ListResourceResponse{} // Stream{}

	logging.FrameworkTrace(ctx, "Calling provider defined ListResource")
	listResource.ListResource(ctx, req, stream)
	logging.FrameworkTrace(ctx, "Called provider defined ListResource")

	if stream.Diagnostics.HasError() {
		// fwStream.Results = slices.Values([]ListResourceEvent{})
		fwStream.Diagnostics = stream.Diagnostics
		return
	}

	if stream.Results == nil {
		// If the provider returned a nil results stream, we treat it as an empty stream.
		stream.Results = iter.Seq[list.ListResourceEvent](func(yield func(list.ListResourceEvent) bool) {})
	}

	fwStream.Results = listResourceEventStreamAdapter(stream.Results)
}

func listResourceEventStreamAdapter(stream iter.Seq[list.ListResourceEvent]) iter.Seq[ListResourceEvent] {
	return func(yieldFw func(ListResourceEvent) bool) {
		yield := func(event list.ListResourceEvent) bool {
			return yieldFw(ListResourceEvent(event))
		}
		stream(yield)
	}
}
