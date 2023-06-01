// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithConfigure{}
var _ resource.ResourceWithConfigure = &ResourceWithConfigure{}

// Declarative resource.ResourceWithConfigure for unit testing.
type ResourceWithConfigure struct {
	*Resource

	// ResourceWithConfigure interface methods
	ConfigureMethod func(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)
}

// Configure satisfies the resource.ResourceWithConfigure interface.
func (r *ResourceWithConfigure) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}
