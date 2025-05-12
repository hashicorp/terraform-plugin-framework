// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// UpgradeIdentity satisfies the tfprotov6.ProviderServer interface.
func (s *Server) UpgradeIdentity(ctx context.Context, proto6Req *tfprotov6.UpgradeResourceIdentityRequest) (*tfprotov6.UpgradeResourceIdentityResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.UpgradeIdentityResponse{}

	if proto6Req == nil {
		return toproto6.UpgradeIdentityResponse(ctx, fwResp), nil
	}

	resource, diags := s.FrameworkServer.Resource(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeIdentityResponse(ctx, fwResp), nil
	}

	identitySchema, diags := s.FrameworkServer.ResourceIdentitySchema(ctx, proto6Req.TypeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeIdentityResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.UpgradeIdentityRequest(ctx, proto6Req, resource, identitySchema)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto6.UpgradeIdentityResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.UpgradeIdentity(ctx, fwReq, fwResp)

	return toproto6.UpgradeIdentityResponse(ctx, fwResp), nil
}
