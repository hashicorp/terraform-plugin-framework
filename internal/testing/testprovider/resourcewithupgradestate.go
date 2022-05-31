package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Resource = &ResourceWithUpgradeState{}
var _ tfsdk.ResourceWithUpgradeState = &ResourceWithUpgradeState{}

// Declarative tfsdk.ResourceWithUpgradeState for unit testing.
type ResourceWithUpgradeState struct {
	*Resource

	// ResourceWithUpgradeState interface methods
	UpgradeStateMethod func(context.Context) map[int64]tfsdk.ResourceStateUpgrader
}

// UpgradeState satisfies the tfsdk.ResourceWithUpgradeState interface.
func (p *ResourceWithUpgradeState) UpgradeState(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
	if p.UpgradeStateMethod == nil {
		return nil
	}

	return p.UpgradeStateMethod(ctx)
}
