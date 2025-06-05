package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

type ResourceSchemaNotFoundError struct {
	TypeName string
}

func (e *ResourceSchemaNotFoundError) Error() string {
	return "resource schema not found for type: " + e.TypeName
}
func (e *ResourceSchemaNotFoundError) Is(err error) bool {
	compatibleErr, ok := err.(*ResourceSchemaNotFoundError)
	if !ok {
		return false
	}

	return e.TypeName == compatibleErr.TypeName
}

type ResourceIdentitySchemaNotFoundError struct {
	TypeName string
}

func (e *ResourceIdentitySchemaNotFoundError) Error() string {
	return "resource identity schema not found for type: " + e.TypeName
}

func (e *ResourceIdentitySchemaNotFoundError) Is(err error) bool {
	compatibleErr, ok := err.(*ResourceIdentitySchemaNotFoundError)
	if !ok {
		return false
	}

	return e.TypeName == compatibleErr.TypeName
}

type ListResourceSchemaNotFoundError struct {
	TypeName string
}

func (e *ListResourceSchemaNotFoundError) Error() string {
	return "list resource schema not found for type: " + e.TypeName
}
func (e *ListResourceSchemaNotFoundError) Is(err error) bool {
	compatibleErr, ok := err.(*ListResourceSchemaNotFoundError)
	if !ok {
		return false
	}

	return e.TypeName == compatibleErr.TypeName
}

type ListResourceConfigError struct {
	TypeName    string
	Diagnostics diag.Diagnostics
}

func (e *ListResourceConfigError) Error() string {
	return "list resource config error for type: " + e.TypeName // + ": " + e.Diagnostics.Error()
}

func (e *ListResourceConfigError) Is(err error) bool {
	compatibleErr, ok := err.(*ListResourceConfigError)
	if !ok {
		return false
	}

	return e.TypeName != compatibleErr.TypeName
}

func (s *Server) ListResource(ctx context.Context, proto6Req *tfprotov6.ListResourceRequest) (*tfprotov6.ListResourceServerStream, error) {
	listResource, err := s.FrameworkServer.ListResourceOrError(ctx, proto6Req.TypeName)
	if err != nil {
		proto6Stream := &tfprotov6.ListResourceServerStream{}
		proto6Stream.Results = func(func(tfprotov6.ListResourceResult) bool) {}

		return proto6Stream, err
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)
	if diags.HasError() {
		proto6Stream := &tfprotov6.ListResourceServerStream{}
		proto6Stream.Results = func(func(tfprotov6.ListResourceResult) bool) {}
		return proto6Stream, &ResourceSchemaNotFoundError{TypeName: proto6Req.TypeName}
	}

	identitySchema, diags := s.FrameworkServer.ResourceIdentitySchema(ctx, proto6Req.TypeName)
	if diags.HasError() {
		proto6Stream := &tfprotov6.ListResourceServerStream{}
		proto6Stream.Results = func(func(tfprotov6.ListResourceResult) bool) {}
		return proto6Stream, &ResourceIdentitySchemaNotFoundError{TypeName: proto6Req.TypeName} // TODO: return a diagnostic instead of an error?
	}

	listResourceSchema, diags := s.FrameworkServer.ListResourceSchema(ctx, proto6Req.TypeName)
	if diags.HasError() {
		proto6Stream := &tfprotov6.ListResourceServerStream{}
		proto6Stream.Results = func(func(tfprotov6.ListResourceResult) bool) {}
		return proto6Stream, &ListResourceSchemaNotFoundError{TypeName: proto6Req.TypeName}
	}

	config, diags := fromproto6.Config(ctx, proto6Req.Config, listResourceSchema)
	if diags.HasError() {
		proto6Stream := &tfprotov6.ListResourceServerStream{}
		proto6Stream.Results = func(func(tfprotov6.ListResourceResult) bool) {}
		return proto6Stream, &ListResourceConfigError{TypeName: proto6Req.TypeName, Diagnostics: diags}
	}

	req := &fwserver.ListRequest{
		Config:                 *config,
		ListResource:           listResource,
		ResourceSchema:         resourceSchema,
		ResourceIdentitySchema: identitySchema,
		IncludeResource:        proto6Req.IncludeResource,
	}
	stream := &fwserver.ListResultsStream{}

	err = s.FrameworkServer.ListResource(ctx, req, stream)
	if err != nil {
		return nil, err
	}

	proto6Stream := &tfprotov6.ListResourceServerStream{}
	proto6Stream.Results = func(push func(tfprotov6.ListResourceResult) bool) {
		for result := range stream.Results {
			var proto6Result tfprotov6.ListResourceResult
			if req.IncludeResource {
				proto6Result = toproto6.ListResourceResultWithResource(ctx, &result)
			} else {
				proto6Result = toproto6.ListResourceResult(ctx, &result)
			}

			if !push(proto6Result) {
				return
			}
		}
	}
	return proto6Stream, nil
}
