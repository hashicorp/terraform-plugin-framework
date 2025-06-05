package proto5server

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func (s *Server) ListResource(ctx context.Context, proto5Req *tfprotov5.ListResourceRequest) (*tfprotov5.ListResourceServerStream, error) {
	listResource, err := s.FrameworkServer.ListResourceOrError(ctx, proto5Req.TypeName)
	if err != nil {
		return nil, err
	}

	listResourceSchema, diags := s.FrameworkServer.ListResourceSchema(ctx, proto5Req.TypeName)
	if diags.HasError() {
		return nil, errors.New("failed to get list resource schema")
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto5Req.TypeName)
	if diags.HasError() {
		return nil, errors.New("failed to get resource schema")
	}

	resourceIdentitySchema, diags := s.FrameworkServer.ResourceIdentitySchema(ctx, proto5Req.TypeName)
	if diags.HasError() {
		return nil, errors.New("failed to get resource identity schema")
	}

	// TODO: safety check on Config
	config, diags := fromproto5.Config(ctx, proto5Req.Config, listResourceSchema)
	if diags.HasError() {
		return nil, errors.New("failed to convert config from proto5 to framework")
	}

	// TODO: this asplodes when no list resource schema is defined

	req := &fwserver.ListRequest{
		ListResource:           listResource,
		Config:                 *config, // TODO: mayBe // this can explode with null input btW
		IncludeResource:        proto5Req.IncludeResource,
		ResourceSchema:         resourceSchema,
		ResourceIdentitySchema: resourceIdentitySchema,
	}

	stream := &fwserver.ListResultsStream{}
	err = s.FrameworkServer.ListResource(ctx, req, stream)
	if err != nil {
		return nil, err
	}

	proto5Stream := &tfprotov5.ListResourceServerStream{}
	proto5Stream.Results = func(push func(tfprotov5.ListResult) bool) {
		for result := range stream.Results {
			identity, diags := toproto5.ResourceIdentity(ctx, result.Identity)
			if diags.HasError() {
				if !push(tfprotov5.ListResult{Diagnostics: toproto5.Diagnostics(ctx, diags)}) {
					return
				}
			}

			resource, diags := toproto5.Resource(ctx, result.Resource)
			if diags.HasError() {
				if !push(tfprotov5.ListResult{Diagnostics: toproto5.Diagnostics(ctx, diags)}) {
					return
				}
			}

			if !push(tfprotov5.ListResult{
				Identity:    identity,
				Resource:    resource,
				DisplayName: result.DisplayName,
				Diagnostics: toproto5.Diagnostics(ctx, result.Diagnostics),
			}) {
				return
			}
		}
	}

	return proto5Stream, nil
}

func (s *Server) ValidateListResourceConfig(ctx context.Context, proto5Req *tfprotov5.ValidateListResourceConfigRequest) (*tfprotov5.ValidateListResourceConfigResponse, error) {
	// This function is intentionally left blank in this example.
	// The implementation would be similar to the one in the original code snippet.
	return nil, nil
}
