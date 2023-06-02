// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSourceWithConfigValidators{}
var _ datasource.DataSourceWithConfigValidators = &DataSourceWithConfigValidators{}

// Declarative datasource.DataSourceWithConfigValidators for unit testing.
type DataSourceWithConfigValidators struct {
	*DataSource

	// DataSourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []datasource.ConfigValidator
}

// ConfigValidators satisfies the datasource.DataSourceWithConfigValidators interface.
func (p *DataSourceWithConfigValidators) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
