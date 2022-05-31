package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Resource = &ResourceWithConfigValidators{}
var _ tfsdk.ResourceWithConfigValidators = &ResourceWithConfigValidators{}

// Declarative tfsdk.ResourceWithConfigValidators for unit testing.
type ResourceWithConfigValidators struct {
	*Resource

	// ResourceWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []tfsdk.ResourceConfigValidator
}

// ConfigValidators satisfies the tfsdk.ResourceWithConfigValidators interface.
func (p *ResourceWithConfigValidators) ConfigValidators(ctx context.Context) []tfsdk.ResourceConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
