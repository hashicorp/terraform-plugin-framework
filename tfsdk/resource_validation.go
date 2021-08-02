package tfsdk

import (
	"context"
)

// ResourceConfigValidator describes reusable Resource configuration validation functionality.
type ResourceConfigValidator interface {
	// Description describes the validation in plain text formatting.
	//
	// This information may be automatically added to resource plain text
	// descriptions by external tooling.
	Description(context.Context) string

	// MarkdownDescription describes the validation in Markdown formatting.
	//
	// This information may be automatically added to resource Markdown
	// descriptions by external tooling.
	MarkdownDescription(context.Context) string

	// Validate performs the validation.
	Validate(context.Context, ValidateResourceConfigRequest, *ValidateResourceConfigResponse)
}

// ResourceWithConfigValidators is an interface type that extends Resource to include declarative validations.
type ResourceWithConfigValidators interface {
	Resource

	// ConfigValidators returns a list of functions which will all be performed during validation.
	ConfigValidators(context.Context) []ResourceConfigValidator
}

// ResourceWithValidateConfig is an interface type that extends Resource to include imperative validation.
type ResourceWithValidateConfig interface {
	Resource

	// ValidateConfig performs the validation.
	ValidateConfig(context.Context, ValidateResourceConfigRequest, *ValidateResourceConfigResponse)
}
