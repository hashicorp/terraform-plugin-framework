package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ApplyResourceChangeResponse returns the *tfprotov5.ApplyResourceChangeResponse
// equivalent of a *fwserver.ApplyResourceChangeResponse.
func ApplyResourceChangeResponse(ctx context.Context, fw *fwserver.ApplyResourceChangeResponse) *tfprotov5.ApplyResourceChangeResponse {
	if fw == nil {
		return nil
	}

	proto5 := &tfprotov5.ApplyResourceChangeResponse{
		Diagnostics: Diagnostics(fw.Diagnostics),
		Private:     fw.Private,
	}

	newState, diags := State(ctx, fw.NewState)

	proto5.Diagnostics = append(proto5.Diagnostics, Diagnostics(diags)...)
	proto5.NewState = newState

	return proto5
}
