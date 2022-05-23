package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ReadResourceResponse returns the *tfprotov6.ReadResourceResponse
// equivalent of a *fwserver.ReadResourceResponse.
func ReadResourceResponse(ctx context.Context, fw *fwserver.ReadResourceResponse) *tfprotov6.ReadResourceResponse {
	if fw == nil {
		return nil
	}

	proto6 := &tfprotov6.ReadResourceResponse{
		Diagnostics: Diagnostics(fw.Diagnostics),
		Private:     fw.Private,
	}

	newState, diags := State(ctx, fw.NewState)

	proto6.Diagnostics = append(proto6.Diagnostics, Diagnostics(diags)...)
	proto6.NewState = newState

	return proto6
}
