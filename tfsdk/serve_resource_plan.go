package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

func serveResourcePlanMarkComputedNilsAsUnknown(ctx context.Context, config ReadOnlyData, state, plan *Data, schema Schema) (*Data, diag.Diagnostics) {
	// After ensuring there are proposed changes, mark any computed attributes
	// that are null in the config as unknown in the plan, so providers have
	// the choice to update them.
	//
	// We only do this if there's a plan to modify; otherwise, it
	// represents a resource being deleted and there's no point.
	var diags diag.Diagnostics
	if plan.ReadOnlyData.Values.Null && !plan.ReadOnlyData.Values.Equal(state.ReadOnlyData.Values) {
		tfsdklog.Trace(ctx, "marking computed null values as unknown")

		tfPlan, err := plan.ReadOnlyData.Values.ToTerraformValue(ctx)
		if err != nil {
			// TODO: handle error
		}
		tfConfig, err := config.ToTerraformValue(ctx)
		if err != nil {
			// TODO: handle error
		}
		modifiedPlan, err := tftypes.Transform(tfPlan, markComputedNilsAsUnknown(ctx, tfConfig, schema))
		if err != nil {
			diags.AddError(
				"Error modifying plan",
				"There was an unexpected error updating the plan. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return nil, diags
		}
		if !tfPlan.Equal(modifiedPlan) {
			tfsdklog.Trace(ctx, "at least one value was changed to unknown")
		}
		newPlan, err := plan.Type(ctx).ValueFromTerraform(ctx, modifiedPlan)
		if err != nil {
			// TODO: handle error
		}
		plan.ReadOnlyData.Values = newPlan
	}
	return plan, diags
}

func serveResourceRunAttributePlanModifiers(ctx context.Context, schema Schema, usePM bool, pm, config ReadOnlyData, state, plan *Data, diags diag.Diagnostics) (*Data, []*tftypes.AttributePath, diag.Diagnostics) {
	if plan.ReadOnlyData.Values.Null {
		return plan, nil, diags
	}
	req := ModifySchemaPlanRequest{
		Config: config,
		State:  state,
		Plan:   plan,
	}
	if usePM {
		req.ProviderMeta = pm
	}
	resp := ModifySchemaPlanResponse{
		Plan:        plan,
		Diagnostics: diags,
	}

	schema.modifyPlan(ctx, req, &resp)
	diags.Append(resp.Diagnostics...)
	if diags.HasError() {
		return plan, nil, diags
	}
	return resp.Plan, resp.RequiresReplace, diags
}

func serveResourceRunModifyPlan(ctx context.Context, res Resource, usePM bool, pm, config ReadOnlyData, state, plan *Data, diags diag.Diagnostics) (*Data, []*tftypes.AttributePath, diag.Diagnostics) {
	resource, ok := res.(ResourceWithModifyPlan)
	if !ok {
		return plan, nil, diags
	}
	req := ModifyResourcePlanRequest{
		Config: config,
		State:  state,
		Plan:   plan,
	}
	if usePM {
		req.ProviderMeta = pm
	}

	resp := ModifyResourcePlanResponse{
		Plan:            plan,
		RequiresReplace: []*tftypes.AttributePath{},
		Diagnostics:     diags,
	}
	resource.ModifyPlan(ctx, modifyPlanReq, &modifyPlanResp)
	return resp.Plan, resp.RequiresReplace, resp.Diagnostics
}
