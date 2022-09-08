package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ provider.Provider = &Provider{}

// Temporarily implement the new interfaces to ease testing transition
var _ provider.ProviderWithDataSources = &Provider{}
var _ provider.ProviderWithResources = &Provider{}

// Declarative provider.Provider for unit testing.
type Provider struct {
	// Provider interface methods
	ConfigureMethod func(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse)
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// ProviderWithDataSources interface methods
	DataSourcesMethod func(context.Context) []func() datasource.DataSource

	// ProviderWithResources interface methods
	ResourcesMethod func(context.Context) []func() resource.Resource
}

// GetSchema satisfies the provider.Provider interface.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	if p == nil || p.ConfigureMethod == nil {
		return
	}

	p.ConfigureMethod(ctx, req, resp)
}

// DataSources satisfies the provider.ProviderWithDataSources interface.
func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	if p == nil || p.DataSourcesMethod == nil {
		return nil
	}

	return p.DataSourcesMethod(ctx)
}

// GetSchema satisfies the provider.Provider interface.
func (p *Provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if p == nil || p.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return p.GetSchemaMethod(ctx)
}

// Resources satisfies the provider.ProviderWithResources interface.
func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	if p == nil || p.ResourcesMethod == nil {
		return nil
	}

	return p.ResourcesMethod(ctx)
}
