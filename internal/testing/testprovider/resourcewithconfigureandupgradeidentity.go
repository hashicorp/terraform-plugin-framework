// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithConfigureAndUpgradeResourceIdentity{}
var _ resource.ResourceWithConfigure = &ResourceWithConfigureAndUpgradeResourceIdentity{}
var _ resource.ResourceWithUpgradeIdentity = &ResourceWithConfigureAndUpgradeResourceIdentity{}

// Declarative resource.ResourceWithConfigureAndUpgradeResourceIdentity for unit testing.
type ResourceWithConfigureAndUpgradeResourceIdentity struct {
	*Resource

	// ResourceWithConfigureAndUpgradeResourceIdentity interface methods
	ConfigureMethod func(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)

	// ResourceWithUpgradeResourceIdentity interface methods
	UpgradeResourceIdentityMethod func(context.Context) map[int64]resource.IdentityUpgrader
}

// Configure satisfies the resource.ResourceWithConfigureAndUpgradeResourceIdentity interface.
func (r *ResourceWithConfigureAndUpgradeResourceIdentity) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}

// UpgradeResourceIdentity satisfies the resource.ResourceWithUpgradeResourceIdentity interface.
func (r *ResourceWithConfigureAndUpgradeResourceIdentity) UpgradeIdentity(ctx context.Context) map[int64]resource.IdentityUpgrader {
	if r.UpgradeResourceIdentityMethod == nil {
		return nil
	}

	return r.UpgradeResourceIdentityMethod(ctx)
}
