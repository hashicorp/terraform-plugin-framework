// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"errors"
	"iter"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ListRequest is the framework server request for the ListResource RPC.
type ListRequest struct {
	// ListResource is an instance of the provider's list resource
	// implementation for a specific managed resource type.
	ListResource list.ListResource

	// Config is the configuration the user supplied for listing resource
	// instances.
	Config *tfsdk.Config

	// IncludeResource indicates whether the provider should populate the
	// Resource field in the ListResult struct.
	IncludeResource bool

	ResourceSchema         fwschema.Schema
	ResourceIdentitySchema fwschema.Schema

	ResourceSchemaType         attr.Type
	ResourceIdentitySchemaType attr.Type
}

// ListResultsStream represents a streaming response to a [ListRequest].  An
// instance of this struct is supplied as an argument to the provider's List
// function. The provider should set a Results iterator function that pushes
// zero or more results of type [ListResult].
//
// For convenience, a provider implementation may choose to convert a slice of
// results into an iterator using [slices.Values].
type ListResultsStream struct {
	// Results is a function that emits [ListResult] values via its push
	// function argument.
	Results iter.Seq[ListResult]
}

func ListResultError(summary string, detail string) ListResult {
	return ListResult{
		Diagnostics: diag.Diagnostics{
			diag.NewErrorDiagnostic(summary, detail),
		},
	}
}

// ListResult represents a listed managed resource instance.
type ListResult struct {
	// Identity is the identity of the managed resource instance. A nil value
	// will raise an error diagnostic.
	Identity *tfsdk.ResourceIdentity

	// Resource is the provider's representation of the attributes of the
	// listed managed resource instance.
	//
	// If [ListRequest.IncludeResource] is true, a nil value will raise
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
func (s *Server) ListResource(ctx context.Context, fwReq *ListRequest, fwStream *ListResultsStream) error {
	listResource := fwReq.ListResource

	if fwReq.Config == nil {
		return errors.New("Invalid ListResource request: Config cannot be nil")
	}

	req := list.ListRequest{
		Config:                     *fwReq.Config,
		IncludeResource:            fwReq.IncludeResource,
		ResourceSchema:             fwReq.ResourceSchema,
		ResourceIdentitySchema:     fwReq.ResourceIdentitySchema,
		ResourceSchemaType:         fwReq.ResourceSchemaType,
		ResourceIdentitySchemaType: fwReq.ResourceIdentitySchemaType,
	}

	stream := &list.ListResultsStream{}

	logging.FrameworkTrace(ctx, "Calling provider defined ListResource")
	listResource.List(ctx, req, stream)
	logging.FrameworkTrace(ctx, "Called provider defined ListResource")

	// If the provider returned a nil results stream, we return an empty stream.
	if stream.Results == nil {
		stream.Results = list.NoListResults
	}

	fwStream.Results = processListResults(req, stream.Results)
	return nil
}

func processListResults(req list.ListRequest, stream iter.Seq[list.ListResult]) iter.Seq[ListResult] {
	return func(push func(ListResult) bool) {
		for result := range stream {
			if !push(processListResult(req, result)) {
				return
			}
		}
	}
}

// processListResult validates the content of a list.ListResult and returns a
// ListResult
func processListResult(req list.ListRequest, result list.ListResult) ListResult {
	if result.Diagnostics.HasError() {
		return ListResult(result)
	}

	if result.Identity == nil { // TODO: is result.Identity.Raw.IsNull() a practical concern?
		return ListResultError(
			"Incomplete List Result",
			"The provider did not populate the Identity field in the ListResourceResult. This may be due to an error in the provider's implementation.",
		)
	}

	if req.IncludeResource {
		if result.Resource == nil { // TODO: is result.Resource.Raw.IsNull() a practical concern?
			result.Diagnostics.AddWarning(
				"Incomplete List Result",
				"The provider did not populate the Resource field in the ListResourceResult. This may be due to the provider not supporting this functionality or an error in the provider's implementation.",
			)
		}
	}

	return ListResult(result) // TODO: do we want to .Copy() the raw Identity and Resource values?
}
