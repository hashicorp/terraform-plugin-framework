// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSourceWithValidateConfig{}
var _ datasource.DataSourceWithValidateConfig = &DataSourceWithValidateConfig{}

// Declarative datasource.DataSourceWithValidateConfig for unit testing.
type DataSourceWithValidateConfig struct {
	*DataSource

	// DataSourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, datasource.ValidateConfigRequest, *datasource.ValidateConfigResponse)
}

// ValidateConfig satisfies the datasource.DataSourceWithValidateConfig interface.
func (p *DataSourceWithValidateConfig) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
