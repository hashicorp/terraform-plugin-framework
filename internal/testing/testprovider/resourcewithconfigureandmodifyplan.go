// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithConfigureAndModifyPlan{}
var _ resource.ResourceWithConfigure = &ResourceWithConfigureAndModifyPlan{}
var _ resource.ResourceWithModifyPlan = &ResourceWithConfigureAndModifyPlan{}

// Declarative resource.ResourceWithConfigureAndModifyPlan for unit testing.
type ResourceWithConfigureAndModifyPlan struct {
	*Resource

	// ResourceWithConfigureAndModifyPlan interface methods
	ConfigureMethod func(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)

	// ResourceWithModifyPlan interface methods
	ModifyPlanMethod func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)
}

// Configure satisfies the resource.ResourceWithConfigureAndModifyPlan interface.
func (r *ResourceWithConfigureAndModifyPlan) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}

// ModifyPlan satisfies the resource.ResourceWithModifyPlan interface.
func (r *ResourceWithConfigureAndModifyPlan) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if r.ModifyPlanMethod == nil {
		return
	}

	r.ModifyPlanMethod(ctx, req, resp)
}
