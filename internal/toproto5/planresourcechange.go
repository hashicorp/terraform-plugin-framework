package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// PlanResourceChangeResponse returns the *tfprotov5.PlanResourceChangeResponse
// equivalent of a *fwserver.PlanResourceChangeResponse.
func PlanResourceChangeResponse(ctx context.Context, fw *fwserver.PlanResourceChangeResponse) *tfprotov5.PlanResourceChangeResponse {
	if fw == nil {
		return nil
	}

	proto5 := &tfprotov5.PlanResourceChangeResponse{
		Diagnostics:    Diagnostics(fw.Diagnostics),
		PlannedPrivate: fw.PlannedPrivate,
	}

	plannedState, diags := State(ctx, fw.PlannedState)

	proto5.Diagnostics = append(proto5.Diagnostics, Diagnostics(diags)...)
	proto5.PlannedState = plannedState
	proto5.RequiresReplace = fw.RequiresReplace

	return proto5
}
