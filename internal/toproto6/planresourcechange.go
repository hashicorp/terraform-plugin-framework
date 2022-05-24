package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// PlanResourceChangeResponse returns the *tfprotov6.PlanResourceChangeResponse
// equivalent of a *fwserver.PlanResourceChangeResponse.
func PlanResourceChangeResponse(ctx context.Context, fw *fwserver.PlanResourceChangeResponse) *tfprotov6.PlanResourceChangeResponse {
	if fw == nil {
		return nil
	}

	proto6 := &tfprotov6.PlanResourceChangeResponse{
		Diagnostics:    Diagnostics(fw.Diagnostics),
		PlannedPrivate: fw.PlannedPrivate,
	}

	plannedState, diags := State(ctx, fw.PlannedState)

	proto6.Diagnostics = append(proto6.Diagnostics, Diagnostics(diags)...)
	proto6.PlannedState = plannedState
	proto6.RequiresReplace = fw.RequiresReplace

	return proto6
}
