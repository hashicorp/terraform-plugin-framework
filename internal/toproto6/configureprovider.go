package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ConfigureProviderResponse returns the *tfprotov6.ConfigureProviderResponse
// equivalent of a *fwserver.ConfigureProviderResponse.
func ConfigureProviderResponse(ctx context.Context, fw *tfsdk.ConfigureProviderResponse) *tfprotov6.ConfigureProviderResponse {
	if fw == nil {
		return nil
	}

	proto6 := &tfprotov6.ConfigureProviderResponse{
		Diagnostics: Diagnostics(fw.Diagnostics),
	}

	return proto6
}
