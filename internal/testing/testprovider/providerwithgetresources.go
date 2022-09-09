package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ provider.Provider = &ProviderWithGetResources{}
var _ provider.ProviderWithGetResources = &ProviderWithGetResources{} //nolint:staticcheck // Internal usage

// Declarative provider.ProviderWithGetResources for unit testing.
type ProviderWithGetResources struct {
	*Provider

	// ProviderWithGetResources interface methods
	//nolint:staticcheck // Internal implementation
	GetResourcesMethod func(context.Context) (map[string]provider.ResourceType, diag.Diagnostics)
}

// GetResources satisfies the provider.ProviderWithGetResources interface.
//
//nolint:staticcheck // Internal implementation
func (p *ProviderWithGetResources) GetResources(ctx context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	if p.GetResourcesMethod == nil {
		return nil, nil
	}

	return p.GetResourcesMethod(ctx)
}
