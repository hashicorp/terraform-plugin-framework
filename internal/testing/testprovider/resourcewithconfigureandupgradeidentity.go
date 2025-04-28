// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithConfigureAndUpgradeIdentity{}
var _ resource.ResourceWithConfigure = &ResourceWithConfigureAndUpgradeIdentity{}
var _ resource.ResourceWithUpgradeIdentity = &ResourceWithConfigureAndUpgradeIdentity{}

// Declarative resource.ResourceWithConfigureAndUpgradeIdentity for unit testing.
type ResourceWithConfigureAndUpgradeIdentity struct {
	*Resource

	// ResourceWithConfigureAndUpgradeIdentity interface methods
	ConfigureMethod func(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)

	// ResourceWithUpgradeIdentity interface methods
	UpgradeResourceIdentityMethod func(context.Context) map[int64]resource.IdentityUpgrader
}

// Configure satisfies the resource.ResourceWithConfigureAndUpgradeIdentity interface.
func (r *ResourceWithConfigureAndUpgradeIdentity) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}

// UpgradeResourceIdentity satisfies the resource.ResourceWithUpgradeIdentity interface.
func (r *ResourceWithConfigureAndUpgradeIdentity) UpgradeResourceIdentity(ctx context.Context) map[int64]resource.IdentityUpgrader {
	if r.UpgradeResourceIdentityMethod == nil {
		return nil
	}

	return r.UpgradeResourceIdentityMethod(ctx)
}
