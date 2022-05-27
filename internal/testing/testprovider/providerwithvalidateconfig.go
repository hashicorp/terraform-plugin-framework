package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Provider = &ProviderWithValidateConfig{}
var _ tfsdk.ProviderWithValidateConfig = &ProviderWithValidateConfig{}

// Declarative tfsdk.ProviderWithValidateConfig for unit testing.
type ProviderWithValidateConfig struct {
	*Provider

	// ProviderWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, tfsdk.ValidateProviderConfigRequest, *tfsdk.ValidateProviderConfigResponse)
}

// GetMetaSchema satisfies the tfsdk.ProviderWithValidateConfig interface.
func (p *ProviderWithValidateConfig) ValidateConfig(ctx context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
