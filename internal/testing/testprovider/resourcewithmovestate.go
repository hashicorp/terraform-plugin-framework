// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithMoveState{}
var _ resource.ResourceWithMoveState = &ResourceWithMoveState{}

// Declarative resource.ResourceWithMoveState for unit testing.
type ResourceWithMoveState struct {
	*Resource

	// ResourceWithMoveState interface methods
	MoveStateMethod func(context.Context) []resource.StateMover
}

// MoveState satisfies the resource.ResourceWithMoveState interface.
func (p *ResourceWithMoveState) MoveState(ctx context.Context) []resource.StateMover {
	if p.MoveStateMethod == nil {
		return nil
	}

	return p.MoveStateMethod(ctx)
}
