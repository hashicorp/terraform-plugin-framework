package tfsdk

import (
	"context"
)

// DataSourceConfigValidator describes reusable data source configuration validation functionality.
type DataSourceConfigValidator interface {
	// Description describes the validation in plain text formatting.
	//
	// This information may be automatically added to data source plain text
	// descriptions by external tooling.
	Description(context.Context) string

	// MarkdownDescription describes the validation in Markdown formatting.
	//
	// This information may be automatically added to data source Markdown
	// descriptions by external tooling.
	MarkdownDescription(context.Context) string

	// Validate performs the validation.
	Validate(context.Context, ValidateDataSourceConfigRequest, *ValidateDataSourceConfigResponse)
}

// DataSourceWithConfigValidators is an interface type that extends DataSource to include declarative validations.
//
// Declaring validation using this methodology simplifies implmentation of
// reusable functionality. These also include descriptions, which can be used
// for automating documentation.
//
// Validation will include ConfigValidators and ValidateConfig, if both are
// implemented, in addition to any Attribute or Type validation.
type DataSourceWithConfigValidators interface {
	DataSource

	// ConfigValidators returns a list of DataSourceConfigValidators. Each DataSourceConfigValidator's Validate method will be called when validating the data source.
	ConfigValidators(context.Context) []DataSourceConfigValidator
}

// DataSourceWithValidateConfig is an interface type that extends DataSource to include imperative validation.
//
// Declaring validation using this methodology simplifies one-off
// functionality that typically applies to a single data source. Any
// documentation of this functionality must be manually added into schema
// descriptions.
//
// Validation will include ConfigValidators and ValidateConfig, if both are
// implemented, in addition to any Attribute or Type validation.
type DataSourceWithValidateConfig interface {
	DataSource

	// ValidateConfig performs the validation.
	ValidateConfig(context.Context, ValidateDataSourceConfigRequest, *ValidateDataSourceConfigResponse)
}
