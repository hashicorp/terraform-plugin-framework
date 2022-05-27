package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.DataSource = &DataSourceWithValidateConfig{}
var _ tfsdk.DataSourceWithValidateConfig = &DataSourceWithValidateConfig{}

// Declarative tfsdk.DataSourceWithValidateConfig for unit testing.
type DataSourceWithValidateConfig struct {
	*DataSource

	// DataSourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, tfsdk.ValidateDataSourceConfigRequest, *tfsdk.ValidateDataSourceConfigResponse)
}

// ValidateConfig satisfies the tfsdk.DataSourceWithValidateConfig interface.
func (p *DataSourceWithValidateConfig) ValidateConfig(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
