package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// This file contains temporary types until GetSchema and Metadata are required
// in DataSource.

var _ datasource.DataSource = &DataSourceWithConfigValidatorsAndGetSchemaAndMetadata{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceWithConfigValidatorsAndGetSchemaAndMetadata{}
var _ datasource.DataSourceWithGetSchema = &DataSourceWithConfigValidatorsAndGetSchemaAndMetadata{}
var _ datasource.DataSourceWithMetadata = &DataSourceWithConfigValidatorsAndGetSchemaAndMetadata{}

// Declarative datasource.DataSourceWithGetSchema for unit testing.
type DataSourceWithConfigValidatorsAndGetSchemaAndMetadata struct {
	*DataSource

	// DataSourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []datasource.ConfigValidator

	// DataSourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// DataSourceWithMetadata interface methods
	MetadataMethod func(context.Context, datasource.MetadataRequest, *datasource.MetadataResponse)
}

// ConfigValidators satisfies the datasource.DataSourceWithConfigValidators interface.
func (d *DataSourceWithConfigValidatorsAndGetSchemaAndMetadata) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if d.ConfigValidatorsMethod == nil {
		return nil
	}

	return d.ConfigValidatorsMethod(ctx)
}

// GetSchema satisfies the datasource.DataSourceWithGetSchema interface.
func (d *DataSourceWithConfigValidatorsAndGetSchemaAndMetadata) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return d.GetSchemaMethod(ctx)
}

// Metadata satisfies the datasource.DataSourceWithMetadata interface.
func (d *DataSourceWithConfigValidatorsAndGetSchemaAndMetadata) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}

var _ datasource.DataSource = &DataSourceWithGetSchemaAndMetadata{}
var _ datasource.DataSourceWithGetSchema = &DataSourceWithGetSchemaAndMetadata{}
var _ datasource.DataSourceWithMetadata = &DataSourceWithGetSchemaAndMetadata{}

// Declarative datasource.DataSourceWithGetSchema for unit testing.
type DataSourceWithGetSchemaAndMetadata struct {
	*DataSource

	// DataSourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// DataSourceWithMetadata interface methods
	MetadataMethod func(context.Context, datasource.MetadataRequest, *datasource.MetadataResponse)
}

// GetSchema satisfies the datasource.DataSourceWithGetSchema interface.
func (d *DataSourceWithGetSchemaAndMetadata) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return d.GetSchemaMethod(ctx)
}

// Metadata satisfies the datasource.DataSourceWithMetadata interface.
func (d *DataSourceWithGetSchemaAndMetadata) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}

var _ datasource.DataSource = &DataSourceWithGetSchemaAndMetadataAndValidateConfig{}
var _ datasource.DataSourceWithGetSchema = &DataSourceWithGetSchemaAndMetadataAndValidateConfig{}
var _ datasource.DataSourceWithMetadata = &DataSourceWithGetSchemaAndMetadataAndValidateConfig{}
var _ datasource.DataSourceWithValidateConfig = &DataSourceWithGetSchemaAndMetadataAndValidateConfig{}

// Declarative datasource.DataSourceWithGetSchema for unit testing.
type DataSourceWithGetSchemaAndMetadataAndValidateConfig struct {
	*DataSource

	// DataSourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// DataSourceWithMetadata interface methods
	MetadataMethod func(context.Context, datasource.MetadataRequest, *datasource.MetadataResponse)

	// DataSourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, datasource.ValidateConfigRequest, *datasource.ValidateConfigResponse)
}

// GetSchema satisfies the datasource.DataSourceWithGetSchema interface.
func (d *DataSourceWithGetSchemaAndMetadataAndValidateConfig) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return d.GetSchemaMethod(ctx)
}

// Metadata satisfies the datasource.DataSourceWithMetadata interface.
func (d *DataSourceWithGetSchemaAndMetadataAndValidateConfig) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}

// ValidateConfig satisfies the datasource.DataSourceWithValidateConfig interface.
func (d *DataSourceWithGetSchemaAndMetadataAndValidateConfig) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if d.ValidateConfigMethod == nil {
		return
	}

	d.ValidateConfigMethod(ctx, req, resp)
}
