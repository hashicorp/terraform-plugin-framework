package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Resource = &ResourceWithValidateConfig{}
var _ tfsdk.ResourceWithValidateConfig = &ResourceWithValidateConfig{}

// Declarative tfsdk.ResourceWithValidateConfig for unit testing.
type ResourceWithValidateConfig struct {
	*Resource

	// ResourceWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, tfsdk.ValidateResourceConfigRequest, *tfsdk.ValidateResourceConfigResponse)
}

// ValidateConfig satisfies the tfsdk.ResourceWithValidateConfig interface.
func (p *ResourceWithValidateConfig) ValidateConfig(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
