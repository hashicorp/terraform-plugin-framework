// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSourceWithConfigure{}
var _ datasource.DataSourceWithConfigure = &DataSourceWithConfigure{}

// Declarative datasource.DataSourceWithConfigure for unit testing.
type DataSourceWithConfigure struct {
	*DataSource

	// DataSourceWithConfigure interface methods
	ConfigureMethod func(context.Context, datasource.ConfigureRequest, *datasource.ConfigureResponse)
}

// Configure satisfies the datasource.DataSourceWithConfigure interface.
func (d *DataSourceWithConfigure) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if d.ConfigureMethod == nil {
		return
	}

	d.ConfigureMethod(ctx, req, resp)
}
