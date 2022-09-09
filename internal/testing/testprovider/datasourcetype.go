package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ provider.DataSourceType = &DataSourceType{} //nolint:staticcheck // Internal implementation

// Declarative provider.DataSourceType for unit testing.
type DataSourceType struct {
	// DataSourceType interface methods
	GetSchemaMethod     func(context.Context) (tfsdk.Schema, diag.Diagnostics)
	NewDataSourceMethod func(context.Context, provider.Provider) (datasource.DataSource, diag.Diagnostics)
}

// GetSchema satisfies the provider.DataSourceType interface.
func (t *DataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if t.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return t.GetSchemaMethod(ctx)
}

// NewDataSource satisfies the provider.DataSourceType interface.
func (t *DataSourceType) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	if t.NewDataSourceMethod == nil {
		return nil, nil
	}

	return t.NewDataSourceMethod(ctx, p)
}
