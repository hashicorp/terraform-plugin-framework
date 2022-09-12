package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSourceWithMetadata{}
var _ datasource.DataSourceWithMetadata = &DataSourceWithMetadata{}

// Declarative datasource.DataSourceWithMetadata for unit testing.
type DataSourceWithMetadata struct {
	*DataSource

	// DataSourceWithMetadata interface methods
	MetadataMethod func(context.Context, datasource.MetadataRequest, *datasource.MetadataResponse)
}

// Metadata satisfies the datasource.DataSourceWithMetadata interface.
func (d *DataSourceWithMetadata) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}
