// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithUpgradeIdentity{}
var _ resource.ResourceWithUpgradeIdentity = &ResourceWithUpgradeIdentity{}

// Declarative resource.ResourceWithUpgradeIdentity for unit testing.
type ResourceWithUpgradeIdentity struct {
	*Resource

	// ResourceWithUpgradeIdentity interface methods
	UpgradeIdentityMethod func(context.Context) map[int64]resource.IdentityUpgrader

	// ResourceWithIdentity interface methods
	IdentitySchemaMethod func(context.Context, resource.IdentitySchemaRequest, *resource.IdentitySchemaResponse)
}

// UpgradeIdentity satisfies the resource.ResourceWithUpgradeIdentity interface.
func (p *ResourceWithUpgradeIdentity) UpgradeIdentity(ctx context.Context) map[int64]resource.IdentityUpgrader {
	if p.UpgradeIdentityMethod == nil {
		return nil
	}

	return p.UpgradeIdentityMethod(ctx)
}

// IdentitySchema implements resource.ResourceWithIdentity.
func (p *ResourceWithUpgradeIdentity) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if p.IdentitySchemaMethod == nil {
		return
	}

	p.IdentitySchemaMethod(ctx, req, resp)
}
