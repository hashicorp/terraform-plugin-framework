// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

var _ action.Action = &ActionWithModifyPlan{}
var _ action.ActionWithModifyPlan = &ActionWithModifyPlan{}

// Declarative action.ActionWithModifyPlan for unit testing.
type ActionWithModifyPlan struct {
	*Action

	// ActionWithModifyPlan interface methods
	ModifyPlanMethod func(context.Context, action.ModifyPlanRequest, *action.ModifyPlanResponse)
}

// ModifyPlan satisfies the action.ActionWithModifyPlan interface.
func (p *ActionWithModifyPlan) ModifyPlan(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
	if p.ModifyPlanMethod == nil {
		return
	}

	p.ModifyPlanMethod(ctx, req, resp)
}
