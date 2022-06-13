package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Resource = &ResourceWithModifyPlan{}
var _ tfsdk.ResourceWithModifyPlan = &ResourceWithModifyPlan{}

// Declarative tfsdk.ResourceWithModifyPlan for unit testing.
type ResourceWithModifyPlan struct {
	*Resource

	// ResourceWithModifyPlan interface methods
	ModifyPlanMethod func(context.Context, tfsdk.ModifyResourcePlanRequest, *tfsdk.ModifyResourcePlanResponse)
}

// ModifyPlan satisfies the tfsdk.ResourceWithModifyPlan interface.
func (p *ResourceWithModifyPlan) ModifyPlan(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {
	if p.ModifyPlanMethod == nil {
		return
	}

	p.ModifyPlanMethod(ctx, req, resp)
}
