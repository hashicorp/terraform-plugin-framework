// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// invokeActionErrorDiagnostics returns a value suitable for
// [InvokeActionServerStream.Events]. It yields a single result that contains
// the given error diagnostics.
func invokeActionErrorDiagnostics(ctx context.Context, diags diag.Diagnostics) (*tfprotov6.InvokeActionServerStream, error) {
	return &tfprotov6.InvokeActionServerStream{
		Events: func(push func(tfprotov6.InvokeActionEvent) bool) {
			push(tfprotov6.InvokeActionEvent{
				Type: tfprotov6.CompletedInvokeActionEventType{
					Diagnostics: toproto6.Diagnostics(ctx, diags),
				},
			})
		},
	}, nil
}

// InvokeAction satisfies the tfprotov6.ProviderServer interface.
func (s *Server) InvokeAction(ctx context.Context, proto6Req *tfprotov6.InvokeActionRequest) (*tfprotov6.InvokeActionServerStream, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.InvokeActionResponse{}

	action, diags := s.FrameworkServer.Action(ctx, proto6Req.ActionType)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	actionSchema, diags := s.FrameworkServer.ActionSchema(ctx, proto6Req.ActionType)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	fwReq, diags := fromproto6.InvokeActionRequest(ctx, proto6Req, action, actionSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return invokeActionErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	// TODO:Actions: Create messaging call back for progress updates

	s.FrameworkServer.InvokeAction(ctx, fwReq, fwResp)

	// TODO:Actions: This is a stub implementation, so we aren't currently exposing any streaming mechanism to the developer.
	// That will eventually need to change to send progress events back to Terraform.
	//
	// This logic will likely need to be moved over to the "toproto" package as well.
	protoStream := &tfprotov6.InvokeActionServerStream{
		Events: func(push func(tfprotov6.InvokeActionEvent) bool) {
			push(tfprotov6.InvokeActionEvent{
				Type: tfprotov6.CompletedInvokeActionEventType{
					Diagnostics: toproto6.Diagnostics(ctx, fwResp.Diagnostics),
				},
			})
		},
	}

	return protoStream, nil
}
