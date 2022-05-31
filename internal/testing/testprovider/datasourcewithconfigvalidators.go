package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.DataSource = &DataSourceWithConfigValidators{}
var _ tfsdk.DataSourceWithConfigValidators = &DataSourceWithConfigValidators{}

// Declarative tfsdk.DataSourceWithConfigValidators for unit testing.
type DataSourceWithConfigValidators struct {
	*DataSource

	// DataSourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []tfsdk.DataSourceConfigValidator
}

// ConfigValidators satisfies the tfsdk.DataSourceWithConfigValidators interface.
func (p *DataSourceWithConfigValidators) ConfigValidators(ctx context.Context) []tfsdk.DataSourceConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
