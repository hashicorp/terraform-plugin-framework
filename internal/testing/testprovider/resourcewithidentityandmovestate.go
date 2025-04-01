// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithIdentityAndMoveState{}
var _ resource.ResourceWithIdentity = &ResourceWithIdentityAndMoveState{}
var _ resource.ResourceWithMoveState = &ResourceWithIdentityAndMoveState{}

// Declarative resource.ResourceWithIdentityAndMoveState for unit testing.
type ResourceWithIdentityAndMoveState struct {
	*Resource

	// ResourceWithIdentity interface methods
	IdentitySchemaMethod func(context.Context, resource.IdentitySchemaRequest, *resource.IdentitySchemaResponse)

	// ResourceWithMoveState interface methods
	MoveStateMethod func(context.Context) []resource.StateMover
}

// IdentitySchema implements resource.ResourceWithIdentity.
func (p *ResourceWithIdentityAndMoveState) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if p.IdentitySchemaMethod == nil {
		return
	}

	p.IdentitySchemaMethod(ctx, req, resp)
}

// MoveState satisfies the resource.ResourceWithMoveState interface.
func (r *ResourceWithIdentityAndMoveState) MoveState(ctx context.Context) []resource.StateMover {
	if r.MoveStateMethod == nil {
		return nil
	}

	return r.MoveStateMethod(ctx)
}
