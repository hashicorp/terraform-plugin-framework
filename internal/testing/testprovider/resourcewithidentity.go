// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithIdentity{}
var _ resource.ResourceWithIdentity = &ResourceWithIdentity{}

// Declarative resource.ResourceWithIdentity for unit testing.
type ResourceWithIdentity struct {
	*Resource

	// ResourceWithIdentity interface methods
	IdentitySchemaMethod func(context.Context, resource.IdentitySchemaRequest, *resource.IdentitySchemaResponse)
}

// IdentitySchema implements resource.ResourceWithIdentity.
func (p *ResourceWithIdentity) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if p.IdentitySchemaMethod == nil {
		return
	}

	p.IdentitySchemaMethod(ctx, req, resp)
}
