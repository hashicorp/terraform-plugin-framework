package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Provider = &Provider{}

// Declarative tfsdk.Provider for unit testing.
type Provider struct {
	// Provider interface methods
	ConfigureMethod      func(context.Context, tfsdk.ConfigureProviderRequest, *tfsdk.ConfigureProviderResponse)
	GetDataSourcesMethod func(context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics)
	GetResourcesMethod   func(context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics)
	GetSchemaMethod      func(context.Context) (tfsdk.Schema, diag.Diagnostics)
}

// GetSchema satisfies the tfsdk.Provider interface.
func (p *Provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	if p.ConfigureMethod == nil {
		return
	}

	p.ConfigureMethod(ctx, req, resp)
}

// GetDataSources satisfies the tfsdk.Provider interface.
func (p *Provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	if p.GetDataSourcesMethod == nil {
		return map[string]tfsdk.DataSourceType{}, nil
	}

	return p.GetDataSourcesMethod(ctx)
}

// GetResources satisfies the tfsdk.Provider interface.
func (p *Provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	if p.GetResourcesMethod == nil {
		return map[string]tfsdk.ResourceType{}, nil
	}

	return p.GetResourcesMethod(ctx)
}

// GetSchema satisfies the tfsdk.Provider interface.
func (p *Provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if p.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return p.GetSchemaMethod(ctx)
}
