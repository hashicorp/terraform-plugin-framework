// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithUpgradeResourceIdentity{}
var _ resource.ResourceWithUpgradeIdentity = &ResourceWithUpgradeResourceIdentity{}

// Declarative resource.ResourceWithUpgradeResourceIdentity for unit testing.
type ResourceWithUpgradeResourceIdentity struct {
	*Resource

	// ResourceWithUpgradeResourceIdentity interface methods
	UpgradeResourceIdentityMethod func(context.Context) map[int64]resource.IdentityUpgrader

	// ResourceWithIdentity interface methods
	IdentitySchemaMethod func(context.Context, resource.IdentitySchemaRequest, *resource.IdentitySchemaResponse)
}

// UpgradeResourceIdentity satisfies the resource.ResourceWithUpgradeResourceIdentity interface.
func (p *ResourceWithUpgradeResourceIdentity) UpgradeIdentity(ctx context.Context) map[int64]resource.IdentityUpgrader {
	if p.UpgradeResourceIdentityMethod == nil {
		return nil
	}

	return p.UpgradeResourceIdentityMethod(ctx)
}

// IdentitySchema implements resource.ResourceWithIdentity.
func (p *ResourceWithUpgradeResourceIdentity) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if p.IdentitySchemaMethod == nil {
		return
	}

	p.IdentitySchemaMethod(ctx, req, resp)
}
