// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.ConfigValidator = &DataSourceConfigValidator{}

// Declarative datasource.ConfigValidator for unit testing.
type DataSourceConfigValidator struct {
	// DataSourceConfigValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateDataSourceMethod  func(context.Context, datasource.ValidateConfigRequest, *datasource.ValidateConfigResponse)
}

// Description satisfies the datasource.ConfigValidator interface.
func (v *DataSourceConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the datasource.ConfigValidator interface.
func (v *DataSourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the datasource.ConfigValidator interface.
func (v *DataSourceConfigValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if v.ValidateDataSourceMethod == nil {
		return
	}

	v.ValidateDataSourceMethod(ctx, req, resp)
}
