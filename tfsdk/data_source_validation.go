package tfsdk

import (
	"context"
)

// DataSourceConfigValidator describes reusable data source configuration validation functionality.
type DataSourceConfigValidator interface {
	// Description describes the validation in plain text formatting.
	Description(context.Context) string

	// MarkdownDescription describes the validation in Markdown formatting.
	MarkdownDescription(context.Context) string

	// Validate performs the validation.
	Validate(context.Context, ValidateDataSourceConfigRequest, *ValidateDataSourceConfigResponse)
}

// DataSourceWithConfigValidators is an interface type that extends DataSource to include declarative validations.
type DataSourceWithConfigValidators interface {
	DataSource

	// ConfigValidators returns a list of DataSourceConfigValidators. Each DataSourceConfigValidator's Validate method will be called when validating the data source.
	ConfigValidators(context.Context) []DataSourceConfigValidator
}

// DataSourceWithValidateConfig is an interface type that extends DataSource to include imperative validation.
type DataSourceWithValidateConfig interface {
	DataSource

	// ValidateConfig performs the validation.
	ValidateConfig(context.Context, ValidateDataSourceConfigRequest, *ValidateDataSourceConfigResponse)
}
