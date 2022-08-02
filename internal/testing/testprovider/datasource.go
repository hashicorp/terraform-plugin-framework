package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSource{}

// Declarative datasource.DataSource for unit testing.
type DataSource struct {
	// DataSource interface methods
	ReadMethod func(context.Context, datasource.ReadRequest, *datasource.ReadResponse)
}

// Read satisfies the datasource.DataSource interface.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.ReadMethod == nil {
		return
	}

	d.ReadMethod(ctx, req, resp)
}
