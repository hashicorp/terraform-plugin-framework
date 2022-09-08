package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ provider.Provider = &ProviderWithGetDataSources{}
var _ provider.ProviderWithGetDataSources = &ProviderWithGetDataSources{} //nolint:staticcheck // Internal usage

// Declarative provider.ProviderWithGetDataSources for unit testing.
type ProviderWithGetDataSources struct {
	*Provider

	// ProviderWithGetDataSources interface methods
	GetDataSourcesMethod func(context.Context) (map[string]provider.DataSourceType, diag.Diagnostics)
}

// GetDataSources satisfies the provider.ProviderWithGetDataSources interface.
func (p *ProviderWithGetDataSources) GetDataSources(ctx context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	if p.GetDataSourcesMethod == nil {
		return nil, nil
	}

	return p.GetDataSourcesMethod(ctx)
}
