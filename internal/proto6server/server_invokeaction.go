package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/pkg/errors"

	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
)

func (s *Server) InvokeAction(ctx context.Context, proto6Req *tfprotov6.InvokeActionRequest, proto6Resp *tfprotov6.InvokeActionResponse) error {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	callBackServer := toproto6.NewInvokeActionCallBackServer(proto6Resp.CallbackServer)

	fwResp := &fwserver.InvokeActionResponse{
		CallbackServer: callBackServer,
	}

	fwAction, diags := s.FrameworkServer.Action(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		// TODO: turn diagnostics into error
		return errors.New("Error invoking action")
	}

	actionSchema, diags := s.FrameworkServer.ActionSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		// TODO: turn diagnostics into error
		return errors.New("Error invoking action")
	}

	fwReq, diags := fromproto6.InvokeActionRequest(ctx, proto6Req, fwAction, actionSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		// TODO: turn diagnostics into error
		return errors.New("Error invoking action")
	}

	s.FrameworkServer.InvokeAction(ctx, fwReq, fwResp)

	return nil
}
