package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// This file contains temporary types until GetSchema and TypeName are required
// in DataSource.

var _ datasource.DataSource = &DataSourceWithConfigValidatorsAndGetSchemaAndTypeName{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceWithConfigValidatorsAndGetSchemaAndTypeName{}
var _ datasource.DataSourceWithGetSchema = &DataSourceWithConfigValidatorsAndGetSchemaAndTypeName{}
var _ datasource.DataSourceWithTypeName = &DataSourceWithConfigValidatorsAndGetSchemaAndTypeName{}

// Declarative datasource.DataSourceWithGetSchema for unit testing.
type DataSourceWithConfigValidatorsAndGetSchemaAndTypeName struct {
	*DataSource

	// DataSourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []datasource.ConfigValidator

	// DataSourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// DataSourceWithTypeName interface methods
	TypeNameMethod func(context.Context, datasource.TypeNameRequest, *datasource.TypeNameResponse)
}

// ConfigValidators satisfies the datasource.DataSourceWithConfigValidators interface.
func (d *DataSourceWithConfigValidatorsAndGetSchemaAndTypeName) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if d.ConfigValidatorsMethod == nil {
		return nil
	}

	return d.ConfigValidatorsMethod(ctx)
}

// GetSchema satisfies the datasource.DataSourceWithGetSchema interface.
func (d *DataSourceWithConfigValidatorsAndGetSchemaAndTypeName) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return d.GetSchemaMethod(ctx)
}

// TypeName satisfies the datasource.DataSourceWithTypeName interface.
func (d *DataSourceWithConfigValidatorsAndGetSchemaAndTypeName) TypeName(ctx context.Context, req datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
	if d.TypeNameMethod == nil {
		return
	}

	d.TypeNameMethod(ctx, req, resp)
}

var _ datasource.DataSource = &DataSourceWithGetSchemaAndTypeName{}
var _ datasource.DataSourceWithGetSchema = &DataSourceWithGetSchemaAndTypeName{}
var _ datasource.DataSourceWithTypeName = &DataSourceWithGetSchemaAndTypeName{}

// Declarative datasource.DataSourceWithGetSchema for unit testing.
type DataSourceWithGetSchemaAndTypeName struct {
	*DataSource

	// DataSourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// DataSourceWithTypeName interface methods
	TypeNameMethod func(context.Context, datasource.TypeNameRequest, *datasource.TypeNameResponse)
}

// GetSchema satisfies the datasource.DataSourceWithGetSchema interface.
func (d *DataSourceWithGetSchemaAndTypeName) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return d.GetSchemaMethod(ctx)
}

// TypeName satisfies the datasource.DataSourceWithTypeName interface.
func (d *DataSourceWithGetSchemaAndTypeName) TypeName(ctx context.Context, req datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
	if d.TypeNameMethod == nil {
		return
	}

	d.TypeNameMethod(ctx, req, resp)
}

var _ datasource.DataSource = &DataSourceWithGetSchemaAndTypeNameAndValidateConfig{}
var _ datasource.DataSourceWithGetSchema = &DataSourceWithGetSchemaAndTypeNameAndValidateConfig{}
var _ datasource.DataSourceWithTypeName = &DataSourceWithGetSchemaAndTypeNameAndValidateConfig{}
var _ datasource.DataSourceWithValidateConfig = &DataSourceWithGetSchemaAndTypeNameAndValidateConfig{}

// Declarative datasource.DataSourceWithGetSchema for unit testing.
type DataSourceWithGetSchemaAndTypeNameAndValidateConfig struct {
	*DataSource

	// DataSourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// DataSourceWithTypeName interface methods
	TypeNameMethod func(context.Context, datasource.TypeNameRequest, *datasource.TypeNameResponse)

	// DataSourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, datasource.ValidateConfigRequest, *datasource.ValidateConfigResponse)
}

// GetSchema satisfies the datasource.DataSourceWithGetSchema interface.
func (d *DataSourceWithGetSchemaAndTypeNameAndValidateConfig) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return d.GetSchemaMethod(ctx)
}

// TypeName satisfies the datasource.DataSourceWithTypeName interface.
func (d *DataSourceWithGetSchemaAndTypeNameAndValidateConfig) TypeName(ctx context.Context, req datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
	if d.TypeNameMethod == nil {
		return
	}

	d.TypeNameMethod(ctx, req, resp)
}

// ValidateConfig satisfies the datasource.DataSourceWithValidateConfig interface.
func (d *DataSourceWithGetSchemaAndTypeNameAndValidateConfig) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if d.ValidateConfigMethod == nil {
		return
	}

	d.ValidateConfigMethod(ctx, req, resp)
}
