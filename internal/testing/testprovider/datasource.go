// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSource{}

// Declarative datasource.DataSource for unit testing.
type DataSource struct {
	// DataSource interface methods
	MetadataMethod func(context.Context, datasource.MetadataRequest, *datasource.MetadataResponse)
	ReadMethod     func(context.Context, datasource.ReadRequest, *datasource.ReadResponse)
	SchemaMethod   func(context.Context, datasource.SchemaRequest, *datasource.SchemaResponse)
}

// Metadata satisfies the datasource.DataSource interface.
func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}

// Read satisfies the datasource.DataSource interface.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.ReadMethod == nil {
		return
	}

	d.ReadMethod(ctx, req, resp)
}

// Schema satisfies the datasource.DataSource interface.
func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	if d.SchemaMethod == nil {
		return
	}

	d.SchemaMethod(ctx, req, resp)
}
