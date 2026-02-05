// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithIdentityAndImportState{}
var _ resource.ResourceWithIdentity = &ResourceWithIdentityAndImportState{}
var _ resource.ResourceWithImportState = &ResourceWithIdentityAndImportState{}

// Declarative resource.ResourceWithIdentityAndImportState for unit testing.
type ResourceWithIdentityAndImportState struct {
	*Resource

	// ResourceWithIdentity interface methods
	IdentitySchemaMethod func(context.Context, resource.IdentitySchemaRequest, *resource.IdentitySchemaResponse)

	// ResourceWithImportState interface methods
	ImportStateMethod func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)
}

// IdentitySchema implements resource.ResourceWithIdentity.
func (p *ResourceWithIdentityAndImportState) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if p.IdentitySchemaMethod == nil {
		return
	}

	p.IdentitySchemaMethod(ctx, req, resp)
}

// ImportState satisfies the resource.ResourceWithImportState interface.
func (r *ResourceWithIdentityAndImportState) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.ImportStateMethod == nil {
		return
	}

	r.ImportStateMethod(ctx, req, resp)
}
