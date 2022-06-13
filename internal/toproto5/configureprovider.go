package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ConfigureProviderResponse returns the *tfprotov5.ConfigureProviderResponse
// equivalent of a *fwserver.ConfigureProviderResponse.
func ConfigureProviderResponse(ctx context.Context, fw *tfsdk.ConfigureProviderResponse) *tfprotov5.ConfigureProviderResponse {
	if fw == nil {
		return nil
	}

	proto5 := &tfprotov5.ConfigureProviderResponse{
		Diagnostics: Diagnostics(fw.Diagnostics),
	}

	return proto5
}
