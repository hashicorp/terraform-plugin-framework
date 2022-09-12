package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// A DataSourceType is a type of data source. For each type of data source this
// provider supports, it should define a type implementing DataSourceType and
// return an instance of it in the map returned by Provider.GetDataSources.
//
// Deprecated: Migrate to datasource.DataSource implementation Configure,
// GetSchema, and Metadata methods. Migrate the provider.Provider
// implementation from the GetDataSources method to the DataSources method.
type DataSourceType interface {
	// GetSchema returns the schema for this data source.
	GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// NewDataSource instantiates a new DataSource of this DataSourceType.
	NewDataSource(context.Context, Provider) (datasource.DataSource, diag.Diagnostics)
}
