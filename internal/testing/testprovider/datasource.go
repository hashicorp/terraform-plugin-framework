package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.DataSource = &DataSource{}

// Declarative tfsdk.DataSource for unit testing.
type DataSource struct {
	// DataSource interface methods
	ReadMethod func(context.Context, tfsdk.ReadDataSourceRequest, *tfsdk.ReadDataSourceResponse)
}

// Read satisfies the tfsdk.DataSource interface.
func (d *DataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	if d.ReadMethod == nil {
		return
	}

	d.Read(ctx, req, resp)
}
