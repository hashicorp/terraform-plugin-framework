package proto5server

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto5"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func (s *Server) GenerateResourceConfig(ctx context.Context, proto5Req *tfprotov5.GenerateResourceConfigRequest) (*tfprotov5.GenerateResourceConfigResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.GenerateResourceConfigResponse{}

	var resourceSchema fwschema.Schema
	var err error
	var diags diag.Diagnostics

	if proto5Req.ResourceSchema != nil {
		resourceSchema, err = fromproto5.ResourceSchema(ctx, proto5Req.ResourceSchema)
		if err != nil {
			fmt.Print("halp")
		}
	} else {
		resourceSchema, diags = s.FrameworkServer.ResourceSchema(ctx, proto5Req.TypeName)

		fwResp.Diagnostics.Append(diags...)
	}

	fwReq, diags := fromproto5.GenerateResourceConfigRequest(ctx, proto5Req, resourceSchema)

	fwResp.Diagnostics.Append(diags...)

	s.FrameworkServer.GenerateResourceConfig(ctx, fwReq, fwResp)

	return toproto5.GenerateResourceConfigResponse(ctx, fwResp), nil

}
