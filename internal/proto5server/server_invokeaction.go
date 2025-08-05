// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// invokeActionErrorDiagnostics returns a value suitable for
// [InvokeActionServerStream.Events]. It yields a single result that contains
// the given error diagnostics.
func invokeActionErrorDiagnostics(ctx context.Context, diags diag.Diagnostics) (*tfprotov5.InvokeActionServerStream, error) {
	return &tfprotov5.InvokeActionServerStream{
		Events: func(push func(tfprotov5.InvokeActionEvent) bool) {
			push(tfprotov5.InvokeActionEvent{
				Type: tfprotov5.CompletedInvokeActionEventType{
					Diagnostics: toproto5.Diagnostics(ctx, diags),
				},
			})
		},
	}, nil
}

// InvokeAction satisfies the tfprotov5.ProviderServer interface.
func (s *Server) InvokeAction(ctx context.Context, proto5Req *tfprotov5.InvokeActionRequest) (*tfprotov5.InvokeActionServerStream, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.InvokeActionResponse{}

	action, diags := s.FrameworkServer.Action(ctx, proto5Req.ActionType)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	actionSchema, diags := s.FrameworkServer.ActionSchema(ctx, proto5Req.ActionType)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	lrSchemas := make([]fwschema.Schema, 0)
	lrIdentitySchemas := make([]fwschema.Schema, 0)
	for _, lrType := range actionSchema.LinkedResourceTypes() {
		switch lrType := lrType.(type) {
		case schema.RawV5LinkedResource:
			// Raw linked resources are not stored on this provider server, so we retrieve the schemas from the
			// action definition directly and convert them to framework schemas.
			lrSchema, err := fromproto5.ResourceSchema(ctx, lrType.GetSchema())
			if err != nil {
				fwResp.Diagnostics.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %q linked resource schema from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n%s", lrType.GetTypeName(), err.Error()),
				)

				return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
			}
			lrSchemas = append(lrSchemas, lrSchema)

			lrIdentitySchema, err := fromproto5.IdentitySchema(ctx, lrType.GetIdentitySchema())
			if err != nil {
				fwResp.Diagnostics.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %q linked resource identity schema from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n%s", lrType.GetTypeName(), err.Error()),
				)

				return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
			}
			lrIdentitySchemas = append(lrIdentitySchemas, lrIdentitySchema)
		case schema.RawV6LinkedResource:
			fwResp.Diagnostics.AddError(
				"Invalid Linked Resource Schema",
				fmt.Sprintf("An unexpected error was encountered when converting %[1]q linked resource schema from the protocol type. "+
					"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
					"Please report this to the provider developer:\n\n"+
					"The %[1]q linked resource is a protocol v6 resource but the provider is being served using protocol v5.", lrType.GetTypeName()),
			)

			return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
		default:
			// Any other linked resource type should be stored on the same provider server as the action,
			// so we can just retrieve it via the type name.
			lrSchema, diags := s.FrameworkServer.ResourceSchema(ctx, lrType.GetTypeName())
			if diags.HasError() {
				fwResp.Diagnostics.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %[1]q linked resource data from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"The %[1]q linked resource was not found on the provider server.", lrType.GetTypeName()),
				)

				return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
			}
			lrSchemas = append(lrSchemas, lrSchema)

			lrIdentitySchema, diags := s.FrameworkServer.ResourceIdentitySchema(ctx, lrType.GetTypeName())
			fwResp.Diagnostics.Append(diags...)
			if fwResp.Diagnostics.HasError() {
				// If the resource is found, the identity schema will only return a diagnostic if the provider implementation
				// returns an error from (resource.Resource).IdentitySchema method.
				return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
			}
			lrIdentitySchemas = append(lrIdentitySchemas, lrIdentitySchema)
		}
	}

	fwReq, diags := fromproto5.InvokeActionRequest(ctx, proto5Req, action, actionSchema, lrSchemas, lrIdentitySchemas)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	protoStream := &tfprotov5.InvokeActionServerStream{
		Events: func(push func(tfprotov5.InvokeActionEvent) bool) {
			// Create a channel for framework to receive progress events
			progressChan := make(chan fwserver.InvokeProgressEvent)
			fwResp.ProgressEvents = progressChan

			// Create a channel to be triggered when the invoke action method has finished
			completedChan := make(chan any)
			go func() {
				s.FrameworkServer.InvokeAction(ctx, fwReq, fwResp)
				close(completedChan)
			}()

			for {
				select {
				// Actions can only push one completed event and it's automatically handled by the framework
				// by closing the completed channel above.
				case <-completedChan:
					push(toproto5.CompletedInvokeActionEventType(ctx, fwResp))
					return

				// Actions can push multiple progress events
				case progressEvent := <-fwResp.ProgressEvents:
					if !push(toproto5.ProgressInvokeActionEventType(ctx, progressEvent)) {
						return
					}
				}
			}
		},
	}

	return protoStream, nil
}
