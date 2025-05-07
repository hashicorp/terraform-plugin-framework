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

type PlanActionRequest struct {
	Action action.Action
	Config *tfsdk.Config
	Schema fwschema.Schema
}

type PlanActionResponse struct {
	Diagnostics   diag.Diagnostics
	PlannedConfig *tfsdk.Plan
}

func (s *Server) PlanAction(ctx context.Context, req *PlanActionRequest, resp *PlanActionResponse) {
	if req == nil {
		return
	}

	nullTfValue := tftypes.NewValue(req.Schema.Type().TerraformType(ctx), nil)

	if req.Config == nil {
		req.Config = &tfsdk.Config{
			Raw:    nullTfValue,
			Schema: req.Schema,
		}
	}

	planReq := action.PlanRequest{
		Config: *req.Config,
	}

	resp.PlannedConfig = configToPlan(*req.Config)

	planResp := action.PlanResponse{
		Diagnostics:   resp.Diagnostics,
		PlannedConfig: *resp.PlannedConfig,
	}

	logging.FrameworkTrace(ctx, "Calling provider defined Action Plan")
	req.Action.Plan(ctx, planReq, &planResp)
	logging.FrameworkTrace(ctx, "Called provider defined Action Plan")

	resp.PlannedConfig = &planResp.PlannedConfig
	resp.Diagnostics = planResp.Diagnostics
}

// planToState returns a *tfsdk.State with a copied value from a tfsdk.Plan.
func configToPlan(config tfsdk.Config) *tfsdk.Plan {
	return &tfsdk.Plan{
		Raw:    config.Raw.Copy(),
		Schema: config.Schema,
	}
}
