package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSourceWithTypeName{}
var _ datasource.DataSourceWithTypeName = &DataSourceWithTypeName{}

// Declarative datasource.DataSourceWithTypeName for unit testing.
type DataSourceWithTypeName struct {
	*DataSource

	// DataSourceWithTypeName interface methods
	TypeNameMethod func(context.Context, datasource.TypeNameRequest, *datasource.TypeNameResponse)
}

// TypeName satisfies the datasource.DataSourceWithTypeName interface.
func (d *DataSourceWithTypeName) TypeName(ctx context.Context, req datasource.TypeNameRequest, resp *datasource.TypeNameResponse) {
	if d.TypeNameMethod == nil {
		return
	}

	d.TypeNameMethod(ctx, req, resp)
}
