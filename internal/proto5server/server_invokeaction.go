// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
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

	fwReq, diags := fromproto5.InvokeActionRequest(ctx, proto5Req, action, actionSchema)

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
	protoStream := &tfprotov5.InvokeActionServerStream{
		Events: func(push func(tfprotov5.InvokeActionEvent) bool) {
			push(tfprotov5.InvokeActionEvent{
				Type: tfprotov5.CompletedInvokeActionEventType{
					Diagnostics: toproto5.Diagnostics(ctx, fwResp.Diagnostics),
				},
			})
		},
	}

	return protoStream, nil
}
