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
		case schema.RawV5LinkedResource:
			// schema.RawLinkedResource are not on the same provider server, so we retrieve the schemas from the
			// definition directly.
			lrSchema, err := fromproto5.ResourceSchema(ctx, lrType.Schema())
			if err != nil {
				// TODO:Actions: Add diagnostic and return
			}
			lrSchemas = append(lrSchemas, lrSchema)

			// TODO:Actions: Implement the mapping logic for identity schemas
			//
			// lrIdentitySchema, err := fromproto5.ResourceIdentitySchema(ctx, lrType.IdentitySchema())
			// if err != nil {
			// 	// TODO:Actions: Add diagnostic and return
			// }
			// lrIdentitySchemas = append(lrIdentitySchemas, lrIdentitySchema)
		case schema.RawV6LinkedResource:
			// TODO:Actions: Would it be invalid to use a v6 linked resource in a v5 action? My initial thought is that
			// this would never happen (since the provider must all be the same protocol version at the end of the day to Terraform,
			// and providers can't build actions for other providers), but I can't think of a reason why we couldn't do this?
			//
			// The data is all the same under the hood, but perhaps there are some validations that might break down when attempting to prevent
			// setting data in nested computed attributes? :shrug:
			//
			// We can very easily validate this in the proto5server/proto6server in our type switch, just need to determine if that restriction is reasonable.
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
