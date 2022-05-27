package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Provider = &ProviderWithConfigValidators{}
var _ tfsdk.ProviderWithConfigValidators = &ProviderWithConfigValidators{}

// Declarative tfsdk.ProviderWithConfigValidators for unit testing.
type ProviderWithConfigValidators struct {
	*Provider

	// ProviderWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []tfsdk.ProviderConfigValidator
}

// GetMetaSchema satisfies the tfsdk.ProviderWithConfigValidators interface.
func (p *ProviderWithConfigValidators) ConfigValidators(ctx context.Context) []tfsdk.ProviderConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
