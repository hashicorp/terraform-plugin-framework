// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// PlanAction satisfies the tfprotov5.ProviderServer interface.
func (s *Server) PlanAction(ctx context.Context, proto5Req *tfprotov5.PlanActionRequest) (*tfprotov5.PlanActionResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.PlanActionResponse{}

	action, diags := s.FrameworkServer.Action(ctx, proto5Req.ActionType)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto5.PlanActionResponse(ctx, fwResp), nil
	}

	actionSchema, diags := s.FrameworkServer.ActionSchema(ctx, proto5Req.ActionType)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto5.PlanActionResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto5.PlanActionRequest(ctx, proto5Req, action, actionSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto5.PlanActionResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.PlanAction(ctx, fwReq, fwResp)

	return toproto5.PlanActionResponse(ctx, fwResp), nil
}
