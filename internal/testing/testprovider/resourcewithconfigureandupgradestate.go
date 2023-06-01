// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithConfigureAndUpgradeState{}
var _ resource.ResourceWithConfigure = &ResourceWithConfigureAndUpgradeState{}
var _ resource.ResourceWithUpgradeState = &ResourceWithConfigureAndUpgradeState{}

// Declarative resource.ResourceWithConfigureAndUpgradeState for unit testing.
type ResourceWithConfigureAndUpgradeState struct {
	*Resource

	// ResourceWithConfigureAndUpgradeState interface methods
	ConfigureMethod func(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)

	// ResourceWithUpgradeState interface methods
	UpgradeStateMethod func(context.Context) map[int64]resource.StateUpgrader
}

// Configure satisfies the resource.ResourceWithConfigureAndUpgradeState interface.
func (r *ResourceWithConfigureAndUpgradeState) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}

// UpgradeState satisfies the resource.ResourceWithUpgradeState interface.
func (r *ResourceWithConfigureAndUpgradeState) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if r.UpgradeStateMethod == nil {
		return nil
	}

	return r.UpgradeStateMethod(ctx)
}
