package tfsdk

import (
	"context"
)

// ProviderConfigValidator describes reusable Provider configuration validation functionality.
type ProviderConfigValidator interface {
	// Description describes the validation in plain text formatting.
	Description(context.Context) string

	// MarkdownDescription describes the validation in Markdown formatting.
	MarkdownDescription(context.Context) string

	// Validate performs the validation.
	Validate(context.Context, ValidateProviderConfigRequest, *ValidateProviderConfigResponse)
}

// ProviderWithConfigValidators is an interface type that extends Provider to include declarative validations.
type ProviderWithConfigValidators interface {
	Provider

	// ConfigValidators returns a list of functions which will all be performed during validation.
	ConfigValidators(context.Context) []ProviderConfigValidator
}

// ProviderWithValidateConfig is an interface type that extends Provider to include imperative validation.
type ProviderWithValidateConfig interface {
	Provider

	// ValidateConfig performs the validation.
	ValidateConfig(context.Context, ValidateProviderConfigRequest, *ValidateProviderConfigResponse)
}
