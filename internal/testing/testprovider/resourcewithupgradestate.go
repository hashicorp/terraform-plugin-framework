// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithUpgradeState{}
var _ resource.ResourceWithUpgradeState = &ResourceWithUpgradeState{}

// Declarative resource.ResourceWithUpgradeState for unit testing.
type ResourceWithUpgradeState struct {
	*Resource

	// ResourceWithUpgradeState interface methods
	UpgradeStateMethod func(context.Context) map[int64]resource.StateUpgrader
}

// UpgradeState satisfies the resource.ResourceWithUpgradeState interface.
func (p *ResourceWithUpgradeState) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if p.UpgradeStateMethod == nil {
		return nil
	}

	return p.UpgradeStateMethod(ctx)
}
