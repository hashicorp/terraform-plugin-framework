package datasource

import "context"

// DataSource represents an instance of a data source type. This is the core
// interface that all data sources must implement.
//
// Data sources can optionally implement these additional concepts:
//
//     - Validation: Schema-based via tfsdk.Attribute or entire configuration
//       via DataSourceWithConfigValidators or DataSourceWithValidateConfig.
//
type DataSource interface {
	// Read is called when the provider must read data source values in
	// order to update state. Config values should be read from the
	// ReadRequest and new state values set on the ReadResponse.
	Read(context.Context, ReadRequest, *ReadResponse)
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

	// ConfigValidators returns a list of ConfigValidators. Each ConfigValidator's Validate method will be called when validating the data source.
	ConfigValidators(context.Context) []ConfigValidator
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
	ValidateConfig(context.Context, ValidateConfigRequest, *ValidateConfigResponse)
}
