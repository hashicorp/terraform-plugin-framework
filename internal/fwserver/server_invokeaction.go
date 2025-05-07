package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type InvokeActionRequest struct {
	Schema fwschema.Schema
	Config *tfsdk.Config
	Action action.Action
}

type InvokeActionResponse struct {
	Diagnostics    diag.Diagnostics
	CallbackServer action.InvokeActionCallBackServer
}

func (s *Server) InvokeAction(ctx context.Context, req *InvokeActionRequest, resp *InvokeActionResponse) {
	if req == nil {
		return
	}

	if resp == nil {
		return
	}

	nullTfValue := tftypes.NewValue(req.Schema.Type().TerraformType(ctx), nil)

	if req.Config == nil {
		req.Config = &tfsdk.Config{
			Raw:    nullTfValue,
			Schema: req.Schema,
		}
	}

	invokeReq := action.InvokeRequest{
		Config: *req.Config,
	}

	invokeResp := action.InvokeResponse{
		Diagnostics:    resp.Diagnostics,
		CallbackServer: resp.CallbackServer,
	}

	logging.FrameworkTrace(ctx, "Calling provider defined Action Invoke")
	req.Action.Invoke(ctx, invokeReq, &invokeResp)
	logging.FrameworkTrace(ctx, "Called provider defined Action Invoke")

	resp.Diagnostics = invokeResp.Diagnostics
}
