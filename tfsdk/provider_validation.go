package tfsdk

import (
	"context"
)

// ProviderConfigValidator describes reusable Provider configuration validation functionality.
type ProviderConfigValidator interface {
	// Description describes the validation in plain text formatting.
	//
	// This information may be automatically added to provider plain text
	// descriptions by external tooling.
	Description(context.Context) string

	// MarkdownDescription describes the validation in Markdown formatting.
	//
	// This information may be automatically added to provider Markdown
	// descriptions by external tooling.
	MarkdownDescription(context.Context) string

	// Validate performs the validation.
	Validate(context.Context, ValidateProviderConfigRequest, *ValidateProviderConfigResponse)
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
	ConfigValidators(context.Context) []ProviderConfigValidator
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
	ValidateConfig(context.Context, ValidateProviderConfigRequest, *ValidateProviderConfigResponse)
}
