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

	// GetResources returns a mapping of resource names to type
	// implementations.
	//
	// Conventionally, resource names should each include a prefix of the
	// provider name and an underscore. For example, a provider named
	// "examplecloud" with resources "thing" and "widget" should use
	// "examplecloud_thing" and "examplecloud_widget" as resource names.
	GetResources(context.Context) (map[string]ResourceType, diag.Diagnostics)

	// GetDataSources returns a mapping of data source name to types
	// implementations.
	//
	// Conventionally, data source names should each include a prefix of the
	// provider name and an underscore. For example, a provider named
	// "examplecloud" with data sources "thing" and "widget" should use
	// "examplecloud_thing" and "examplecloud_widget" as data source names.
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
