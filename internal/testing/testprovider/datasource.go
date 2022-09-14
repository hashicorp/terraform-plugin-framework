package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ datasource.DataSource = &DataSource{}

// Declarative datasource.DataSource for unit testing.
type DataSource struct {
	// DataSource interface methods
	MetadataMethod  func(context.Context, datasource.MetadataRequest, *datasource.MetadataResponse)
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)
	ReadMethod      func(context.Context, datasource.ReadRequest, *datasource.ReadResponse)
}

// Metadata satisfies the datasource.DataSource interface.
func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}

// GetSchema satisfies the datasource.DataSource interface.
func (d *DataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if d.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return d.GetSchemaMethod(ctx)
}

// Read satisfies the datasource.DataSource interface.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.ReadMethod == nil {
		return
	}

	d.ReadMethod(ctx, req, resp)
}
