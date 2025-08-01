// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"fmt"

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
			// Raw linked resources are not stored on this provider server, so we retrieve the schemas from the
			// action definition directly and convert them to framework schemas.
			lrSchema, err := fromproto5.ResourceSchema(ctx, lrType.GetSchema())
			if err != nil {
				fwResp.Diagnostics.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %q linked resource schema from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n%s", lrType.GetTypeName(), err.Error()),
				)

				return toproto5.PlanActionResponse(ctx, fwResp), nil //nolint:nilerr // error is assigned to fwResp.Diagnostics
			}
			lrSchemas = append(lrSchemas, lrSchema)

			lrIdentitySchema, err := fromproto5.IdentitySchema(ctx, lrType.GetIdentitySchema())
			if err != nil {
				fwResp.Diagnostics.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %q linked resource identity schema from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n%s", lrType.GetTypeName(), err.Error()),
				)

				return toproto5.PlanActionResponse(ctx, fwResp), nil //nolint:nilerr // error is assigned to fwResp.Diagnostics
			}
			lrIdentitySchemas = append(lrIdentitySchemas, lrIdentitySchema)
		case schema.RawV6LinkedResource:
			fwResp.Diagnostics.AddError(
				"Invalid Linked Resource Schema",
				fmt.Sprintf("An unexpected error was encountered when converting %[1]q linked resource schema from the protocol type. "+
					"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
					"Please report this to the provider developer:\n\n"+
					"The %[1]q linked resource is a protocol v6 resource but the provider is being served using protocol v5.", lrType.GetTypeName()),
			)

			return toproto5.PlanActionResponse(ctx, fwResp), nil //nolint:nilerr // error is assigned to fwResp.Diagnostics
		default:
			// Any other linked resource type should be stored on the same provider server as the action,
			// so we can just retrieve it via the type name.
			lrSchema, diags := s.FrameworkServer.ResourceSchema(ctx, lrType.GetTypeName())
			if diags.HasError() {
				fwResp.Diagnostics.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %[1]q linked resource data from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"The %[1]q linked resource was not found on the provider server.", lrType.GetTypeName()),
				)

				return toproto5.PlanActionResponse(ctx, fwResp), nil
			}
			lrSchemas = append(lrSchemas, lrSchema)

			lrIdentitySchema, diags := s.FrameworkServer.ResourceIdentitySchema(ctx, lrType.GetTypeName())
			fwResp.Diagnostics.Append(diags...)
			if fwResp.Diagnostics.HasError() {
				// If the resource is found, the identity schema will only return a diagnostic if the provider implementation
				// returns an error from (resource.Resource).IdentitySchema method.
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
