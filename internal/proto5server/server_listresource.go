// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ListRequestErrorDiagnostics returns a value suitable for
// [ListResourceServerStream.Results]. It yields a single result that contains
// the given error diagnostics.
func ListRequestErrorDiagnostics(ctx context.Context, diags ...diag.Diagnostic) (*tfprotov5.ListResourceServerStream, error) {
	protoDiags := toproto5.Diagnostics(ctx, diags)
	return &tfprotov5.ListResourceServerStream{
		Results: func(push func(tfprotov5.ListResourceResult) bool) {
			push(tfprotov5.ListResourceResult{Diagnostics: protoDiags})
		},
	}, nil
}

func (s *Server) ListResource(ctx context.Context, protoReq *tfprotov5.ListResourceRequest) (*tfprotov5.ListResourceServerStream, error) {
	protoStream := &tfprotov5.ListResourceServerStream{Results: tfprotov5.NoListResults}
	allDiags := diag.Diagnostics{}

	listResource, diags := s.FrameworkServer.ListResourceType(ctx, protoReq.TypeName)
	allDiags.Append(diags...)
	if diags.HasError() {
		return ListRequestErrorDiagnostics(ctx, allDiags...)
	}

	listResourceSchema, diags := s.FrameworkServer.ListResourceSchema(ctx, protoReq.TypeName)
	allDiags.Append(diags...)
	if diags.HasError() {
		return ListRequestErrorDiagnostics(ctx, allDiags...)
	}

	config, diags := fromproto5.Config(ctx, protoReq.Config, listResourceSchema)
	allDiags.Append(diags...)
	if diags.HasError() {
		return ListRequestErrorDiagnostics(ctx, allDiags...)
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, protoReq.TypeName)
	allDiags.Append(diags...)
	if diags.HasError() {
		return ListRequestErrorDiagnostics(ctx, allDiags...)
	}

	identitySchema, diags := s.FrameworkServer.ResourceIdentitySchema(ctx, protoReq.TypeName)
	allDiags.Append(diags...)
	if diags.HasError() {
		return ListRequestErrorDiagnostics(ctx, allDiags...)
	}

	req := &fwserver.ListRequest{
		Config:                 config,
		ListResource:           listResource,
		ResourceSchema:         resourceSchema,
		ResourceIdentitySchema: identitySchema,
		IncludeResource:        protoReq.IncludeResource,
	}
	stream := &fwserver.ListResultsStream{}

	err := s.FrameworkServer.ListResource(ctx, req, stream)
	if err != nil {
		return protoStream, err
	}

	protoStream.Results = func(push func(tfprotov5.ListResourceResult) bool) {
		for result := range stream.Results {
			var protoResult tfprotov5.ListResourceResult
			if req.IncludeResource {
				protoResult = toproto5.ListResourceResultWithResource(ctx, &result)
			} else {
				protoResult = toproto5.ListResourceResult(ctx, &result)
			}

			if !push(protoResult) {
				return
			}
		}
	}
	return protoStream, nil
}
