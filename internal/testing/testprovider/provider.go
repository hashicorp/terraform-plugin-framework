package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ provider.Provider = &Provider{}

// Declarative provider.Provider for unit testing.
type Provider struct {
	// Provider interface methods
	ConfigureMethod      func(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse)
	GetDataSourcesMethod func(context.Context) (map[string]provider.DataSourceType, diag.Diagnostics)
	GetResourcesMethod   func(context.Context) (map[string]provider.ResourceType, diag.Diagnostics)
	GetSchemaMethod      func(context.Context) (tfsdk.Schema, diag.Diagnostics)
}

// GetSchema satisfies the provider.Provider interface.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	if p.ConfigureMethod == nil {
		return
	}

	p.ConfigureMethod(ctx, req, resp)
}

// GetDataSources satisfies the provider.Provider interface.
func (p *Provider) GetDataSources(ctx context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	if p.GetDataSourcesMethod == nil {
		return map[string]provider.DataSourceType{}, nil
	}

	return p.GetDataSourcesMethod(ctx)
}

// GetResources satisfies the provider.Provider interface.
func (p *Provider) GetResources(ctx context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	if p.GetResourcesMethod == nil {
		return map[string]provider.ResourceType{}, nil
	}

	return p.GetResourcesMethod(ctx)
}

// GetSchema satisfies the provider.Provider interface.
func (p *Provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if p.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return p.GetSchemaMethod(ctx)
}
