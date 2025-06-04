// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &Provider{}

// Declarative provider.Provider for unit testing.
type Provider struct {
	// Provider interface methods
	MetadataMethod           func(context.Context, provider.MetadataRequest, *provider.MetadataResponse)
	ConfigureMethod          func(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse)
	SchemaMethod             func(context.Context, provider.SchemaRequest, *provider.SchemaResponse)
	DataSourcesMethod        func(context.Context) []func() datasource.DataSource
	EphemeralResourcesMethod func(context.Context) []func() ephemeral.EphemeralResource
	ListResourcesMethod      func(context.Context) []func() list.ListResource
	ResourcesMethod          func(context.Context) []func() resource.Resource
}

// Configure satisfies the provider.Provider interface.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	if p == nil || p.ConfigureMethod == nil {
		return
	}

	p.ConfigureMethod(ctx, req, resp)
}

// DataSources satisfies the provider.Provider interface.
func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	if p == nil || p.DataSourcesMethod == nil {
		return nil
	}

	return p.DataSourcesMethod(ctx)
}

// Metadata satisfies the provider.Provider interface.
func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	if p == nil || p.MetadataMethod == nil {
		return
	}

	p.MetadataMethod(ctx, req, resp)
}

// Schema satisfies the provider.Provider interface.
func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	if p == nil || p.SchemaMethod == nil {
		return
	}

	p.SchemaMethod(ctx, req, resp)
}

// ListResources satisfies the provider.Provider interface.
func (p *Provider) ListResources(ctx context.Context) []func() list.ListResource {
	if p == nil || p.ListResourcesMethod == nil {
		return nil
	}

	return p.ListResourcesMethod(ctx)
}

// Resources satisfies the provider.Provider interface.
func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	if p == nil || p.ResourcesMethod == nil {
		return nil
	}

	return p.ResourcesMethod(ctx)
}

func (p *Provider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	if p == nil || p.EphemeralResourcesMethod == nil {
		return nil
	}

	return p.EphemeralResourcesMethod(ctx)
}
