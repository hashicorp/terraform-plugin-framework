package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.DataSourceType = &DataSourceType{}

// Declarative tfsdk.DataSourceType for unit testing.
type DataSourceType struct {
	// DataSourceType interface methods
	GetSchemaMethod     func(context.Context) (tfsdk.Schema, diag.Diagnostics)
	NewDataSourceMethod func(context.Context, tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics)
}

// GetSchema satisfies the tfsdk.DataSourceType interface.
func (t *DataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if t.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return t.GetSchemaMethod(ctx)
}

// NewDataSource satisfies the tfsdk.DataSourceType interface.
func (t *DataSourceType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	if t.NewDataSourceMethod == nil {
		return nil, nil
	}

	return t.NewDataSourceMethod(ctx, p)
}
