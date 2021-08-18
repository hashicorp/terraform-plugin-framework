package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeProviderWithValidateConfig struct {
	*testServeProvider
}

func (t *testServeProviderWithValidateConfig) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

var testServeProviderWithValidateConfigType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"string": tftypes.String,
	},
}

func (p testServeProviderWithValidateConfig) ValidateConfig(ctx context.Context, req ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {
	p.validateProviderConfigImpl(ctx, req, resp)
}
