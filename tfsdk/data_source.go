package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// A DataSourceType is a type of data source. For each type of data source this
// provider supports, it should define a type implementing DataSourceType and
// return an instance of it in the map returned by Provider.GetDataSources.
type DataSourceType interface {
	// GetSchema returns the schema for this data source.
	GetSchema(context.Context) (Schema, diag.Diagnostics)

	// NewDataSource instantiates a new DataSource of this DataSourceType.
	NewDataSource(context.Context, Provider) (DataSource, diag.Diagnostics)
}

// DataSource represents a data source instance. This is the core interface that
// all data sources must implement.
type DataSource interface {
	// Read is called when the provider must read data source values in
	// order to update state. Config values should be read from the
	// ReadDataSourceRequest and new state values set on the
	// ReadDataSourceResponse.
	Read(context.Context, ReadDataSourceRequest, *ReadDataSourceResponse)
}
