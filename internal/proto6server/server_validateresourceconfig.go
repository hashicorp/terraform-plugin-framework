package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ValidateResourceConfig satisfies the tfprotov6.ProviderServer interface.
func (s *Server) ValidateResourceConfig(ctx context.Context, proto6Req *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ValidateResourceConfigResponse{}

	resourceType, diags := s.FrameworkServer.ResourceType(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	resourceSchema, diags := s.FrameworkServer.ResourceSchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.ValidateResourceConfigRequest(ctx, proto6Req, resourceType, resourceSchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.ValidateResourceConfig(ctx, fwReq, fwResp)

	return toproto6.ValidateResourceConfigResponse(ctx, fwResp), nil
}
