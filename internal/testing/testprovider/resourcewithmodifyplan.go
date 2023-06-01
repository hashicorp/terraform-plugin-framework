// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithModifyPlan{}
var _ resource.ResourceWithModifyPlan = &ResourceWithModifyPlan{}

// Declarative resource.ResourceWithModifyPlan for unit testing.
type ResourceWithModifyPlan struct {
	*Resource

	// ResourceWithModifyPlan interface methods
	ModifyPlanMethod func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)
}

// ModifyPlan satisfies the resource.ResourceWithModifyPlan interface.
func (p *ResourceWithModifyPlan) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if p.ModifyPlanMethod == nil {
		return
	}

	p.ModifyPlanMethod(ctx, req, resp)
}
