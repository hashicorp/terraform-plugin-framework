// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

var _ action.Action = &ActionWithConfigureAndModifyPlan{}
var _ action.ActionWithConfigure = &ActionWithConfigureAndModifyPlan{}
var _ action.ActionWithModifyPlan = &ActionWithConfigureAndModifyPlan{}

// Declarative action.ActionWithConfigureAndModifyPlan for unit testing.
type ActionWithConfigureAndModifyPlan struct {
	*Action

	// ActionWithConfigureAndModifyPlan interface methods
	ConfigureMethod func(context.Context, action.ConfigureRequest, *action.ConfigureResponse)

	// ActionWithModifyPlan interface methods
	ModifyPlanMethod func(context.Context, action.ModifyPlanRequest, *action.ModifyPlanResponse)
}

// Configure satisfies the action.ActionWithConfigureAndModifyPlan interface.
func (r *ActionWithConfigureAndModifyPlan) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}

// ModifyPlan satisfies the action.ActionWithModifyPlan interface.
func (r *ActionWithConfigureAndModifyPlan) ModifyPlan(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
	if r.ModifyPlanMethod == nil {
		return
	}

	r.ModifyPlanMethod(ctx, req, resp)
}
