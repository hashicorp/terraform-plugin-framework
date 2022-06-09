package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ReadResourceResponse returns the *tfprotov5.ReadResourceResponse
// equivalent of a *fwserver.ReadResourceResponse.
func ReadResourceResponse(ctx context.Context, fw *fwserver.ReadResourceResponse) *tfprotov5.ReadResourceResponse {
	if fw == nil {
		return nil
	}

	proto5 := &tfprotov5.ReadResourceResponse{
		Diagnostics: Diagnostics(fw.Diagnostics),
		Private:     fw.Private,
	}

	newState, diags := State(ctx, fw.NewState)

	proto5.Diagnostics = append(proto5.Diagnostics, Diagnostics(diags)...)
	proto5.NewState = newState

	return proto5
}
