// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
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

	lrSchemas := make([]fwschema.Schema, 0)
	lrIdentitySchemas := make([]fwschema.Schema, 0)
	for _, lrType := range actionSchema.LinkedResourceTypes() {
		switch lrType := lrType.(type) {
		case schema.RawLinkedResource:
			// schema.RawLinkedResource are not on the same provider server, so we retrieve the schemas from the
			// definition directly.
			lrSchemas = append(lrSchemas, lrType.GetSchema())
			lrIdentitySchemas = append(lrIdentitySchemas, lrType.GetIdentitySchema())
		default:
			// Any other linked resource type should be on the same provider server as the action, so we can just retrieve it
			lrSchema, diags := s.FrameworkServer.ResourceSchema(ctx, lrType.GetTypeName())

			fwResp.Diagnostics.Append(diags...)
			if fwResp.Diagnostics.HasError() {
				// TODO:Actions: Better error message
				return toproto5.PlanActionResponse(ctx, fwResp), nil
			}

			lrSchemas = append(lrSchemas, lrSchema)

			lrIdentitySchema, diags := s.FrameworkServer.ResourceIdentitySchema(ctx, lrType.GetTypeName())

			fwResp.Diagnostics.Append(diags...)
			if fwResp.Diagnostics.HasError() {
				// TODO:Actions: Better error message
				return toproto5.PlanActionResponse(ctx, fwResp), nil
			}

			lrIdentitySchemas = append(lrIdentitySchemas, lrIdentitySchema)
		}
	}

	fwReq, diags := fromproto5.PlanActionRequest(ctx, proto5Req, action, actionSchema, lrSchemas, lrIdentitySchemas)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return toproto5.PlanActionResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.PlanAction(ctx, fwReq, fwResp)

	return toproto5.PlanActionResponse(ctx, fwResp), nil
}
