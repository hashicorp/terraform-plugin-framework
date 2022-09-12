package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ datasource.DataSource = &DataSourceWithGetSchema{}
var _ datasource.DataSourceWithGetSchema = &DataSourceWithGetSchema{}

// Declarative datasource.DataSourceWithGetSchema for unit testing.
type DataSourceWithGetSchema struct {
	*DataSource

	// DataSourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)
}

// GetSchema satisfies the datasource.DataSourceWithGetSchema interface.
func (d *DataSourceWithGetSchema) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return d.GetSchemaMethod(ctx)
}
