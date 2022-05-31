package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.DataSourceConfigValidator = &DataSourceConfigValidator{}

// Declarative tfsdk.DataSourceConfigValidator for unit testing.
type DataSourceConfigValidator struct {
	// DataSourceConfigValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateMethod            func(context.Context, tfsdk.ValidateDataSourceConfigRequest, *tfsdk.ValidateDataSourceConfigResponse)
}

// Description satisfies the tfsdk.DataSourceConfigValidator interface.
func (v *DataSourceConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the tfsdk.DataSourceConfigValidator interface.
func (v *DataSourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the tfsdk.DataSourceConfigValidator interface.
func (v *DataSourceConfigValidator) Validate(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
