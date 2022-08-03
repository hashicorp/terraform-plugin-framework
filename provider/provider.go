package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Provider is the core interface that all Terraform providers must implement.
//
// Providers can optionally implement these additional concepts:
//
//     - Validation: Schema-based via tfsdk.Attribute or entire configuration
//       via ProviderWithConfigValidators or ProviderWithValidateConfig.
//     - Meta Schema: ProviderWithMetaSchema
//
type Provider interface {
	// GetSchema returns the schema for this provider's configuration. If
	// this provider has no configuration, return an empty schema.Schema.
	GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// Configure is called at the beginning of the provider lifecycle, when
	// Terraform sends to the provider the values the user specified in the
	// provider configuration block. These are supplied in the
	// ConfigureProviderRequest argument.
	// Values from provider configuration are often used to initialise an
	// API client, which should be stored on the struct implementing the
	// Provider interface.
	Configure(context.Context, ConfigureRequest, *ConfigureResponse)

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

// ProviderWithConfigValidators is an interface type that extends Provider to include declarative validations.
//
// Declaring validation using this methodology simplifies implementation of
// reusable functionality. These also include descriptions, which can be used
// for automating documentation.
//
// Validation will include ConfigValidators and ValidateConfig, if both are
// implemented, in addition to any Attribute or Type validation.
type ProviderWithConfigValidators interface {
	Provider

	// ConfigValidators returns a list of functions which will all be performed during validation.
	ConfigValidators(context.Context) []ConfigValidator
}

// ProviderWithMetaSchema is a provider with a provider meta schema.
// This functionality is currently experimental and subject to change or break
// without warning; it should only be used by providers that are collaborating
// on its use with the Terraform team.
type ProviderWithMetaSchema interface {
	Provider

	// GetMetaSchema returns the provider meta schema.
	GetMetaSchema(context.Context) (tfsdk.Schema, diag.Diagnostics)
}

// ProviderWithValidateConfig is an interface type that extends Provider to include imperative validation.
//
// Declaring validation using this methodology simplifies one-off
// functionality that typically applies to a single provider. Any documentation
// of this functionality must be manually added into schema descriptions.
//
// Validation will include ConfigValidators and ValidateConfig, if both are
// implemented, in addition to any Attribute or Type validation.
type ProviderWithValidateConfig interface {
	Provider

	// ValidateConfig performs the validation.
	ValidateConfig(context.Context, ValidateConfigRequest, *ValidateConfigResponse)
}
