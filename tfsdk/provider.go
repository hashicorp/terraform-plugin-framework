package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Provider is the core interface that all Terraform providers must implement.
type Provider interface {
	// GetSchema returns the schema for this provider's configuration. If
	// this provider has no configuration, return an empty schema.Schema.
	GetSchema(context.Context) (Schema, diag.Diagnostics)

	// Configure is called at the beginning of the provider lifecycle, when
	// Terraform sends to the provider the values the user specified in the
	// provider configuration block. These are supplied in the
	// ConfigureProviderRequest argument.
	// Values from provider configuration are often used to initialise an
	// API client, which should be stored on the struct implementing the
	// Provider interface.
	Configure(context.Context, ConfigureProviderRequest, *ConfigureProviderResponse)

	// GetResources returns a map of the resource types this provider
	// supports.
	GetResources(context.Context) (map[string]ResourceType, diag.Diagnostics)

	// GetDataSources returns a map of the data source types this provider
	// supports.
	GetDataSources(context.Context) (map[string]DataSourceType, diag.Diagnostics)
}

// ProviderWithProviderMeta is a provider with a provider meta schema.
// This functionality is currently experimental and subject to change or break
// without warning; it should only be used by providers that are collaborating
// on its use with the Terraform team.
type ProviderWithProviderMeta interface {
	Provider
	// GetMetaSchema returns the provider meta schema.
	GetMetaSchema(context.Context) (Schema, diag.Diagnostics)
}
