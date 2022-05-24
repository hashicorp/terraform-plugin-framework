package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// UpgradeResourceState satisfies the tfprotov6.ProviderServer interface.
func (s *Server) UpgradeResourceState(ctx context.Context, proto6Req *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.UpgradeResourceStateResponse{}

	if proto6Req == nil {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.UpgradeResourceStateRequest(ctx, proto6Req, resourceType, resourceSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.UpgradeResourceState(ctx, fwReq, fwResp)

	return toproto6.UpgradeResourceStateResponse(ctx, fwResp), nil
}
