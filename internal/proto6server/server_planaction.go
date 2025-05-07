package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
)

func (s *Server) PlanAction(ctx context.Context, proto6Req *tfprotov6.PlanActionRequest) (*tfprotov6.PlanActionResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.PlanActionResponse{}

	action, diags := s.FrameworkServer.Action(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.PlanActionResponse(ctx, fwResp), nil
	}

	actionSchema, diags := s.FrameworkServer.ActionSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.PlanActionResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.PlanActionRequest(ctx, proto6Req, action, actionSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.PlanActionResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.PlanAction(ctx, fwReq, fwResp)

	return toproto6.PlanActionResponse(ctx, fwResp), nil
}
