package fromproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ConfigureProviderRequest returns the *fwserver.ConfigureProviderRequest
// equivalent of a *tfprotov5.ConfigureProviderRequest.
func ConfigureProviderRequest(ctx context.Context, proto5 *tfprotov5.ConfigureProviderRequest, providerSchema *tfsdk.Schema) (*tfsdk.ConfigureProviderRequest, diag.Diagnostics) {
	if proto5 == nil {
		return nil, nil
	}

	fw := &tfsdk.ConfigureProviderRequest{
		TerraformVersion: proto5.TerraformVersion,
	}

	config, diags := Config(ctx, proto5.Config, providerSchema)

	if config != nil {
		fw.Config = *config
	}

	return fw, diags
}
