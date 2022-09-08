package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
)

var _ datasource.DataSource = &DataSourceWithGetSchema{}
var _ datasource.DataSourceWithGetSchema = &DataSourceWithGetSchema{}

// Declarative datasource.DataSourceWithGetSchema for unit testing.
type DataSourceWithGetSchema struct {
	*DataSource

	// DataSourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (fwschema.Schema, diag.Diagnostics)
}

// GetSchema satisfies the datasource.DataSourceWithGetSchema interface.
func (d *DataSourceWithGetSchema) GetSchema(ctx context.Context) (fwschema.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return nil, nil
	}

	return d.GetSchemaMethod(ctx)
}
