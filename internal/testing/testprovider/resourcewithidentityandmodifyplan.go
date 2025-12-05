// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithIdentityAndModifyPlan{}
var _ resource.ResourceWithIdentity = &ResourceWithIdentityAndModifyPlan{}
var _ resource.ResourceWithModifyPlan = &ResourceWithIdentityAndModifyPlan{}

// Declarative resource.ResourceWithIdentityAndModifyPlan for unit testing.
type ResourceWithIdentityAndModifyPlan struct {
	*Resource

	// ResourceWithIdentity interface methods
	IdentitySchemaMethod func(context.Context, resource.IdentitySchemaRequest, *resource.IdentitySchemaResponse)

	// ResourceWithModifyPlan interface methods
	ModifyPlanMethod func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)
}

// IdentitySchema implements resource.ResourceWithIdentity.
func (p *ResourceWithIdentityAndModifyPlan) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if p.IdentitySchemaMethod == nil {
		return
	}

	p.IdentitySchemaMethod(ctx, req, resp)
}

// ModifyPlan satisfies the resource.ResourceWithModifyPlan interface.
func (r *ResourceWithIdentityAndModifyPlan) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if r.ModifyPlanMethod == nil {
		return
	}

	r.ModifyPlanMethod(ctx, req, resp)
}
