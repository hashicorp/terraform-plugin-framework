// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithConfigureAndImportState{}
var _ resource.ResourceWithConfigure = &ResourceWithConfigureAndImportState{}
var _ resource.ResourceWithImportState = &ResourceWithConfigureAndImportState{}

// Declarative resource.ResourceWithConfigureAndImportState for unit testing.
type ResourceWithConfigureAndImportState struct {
	*Resource

	// ResourceWithConfigureAndImportState interface methods
	ConfigureMethod func(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)

	// ResourceWithImportState interface methods
	ImportStateMethod func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)
}

// Configure satisfies the resource.ResourceWithConfigureAndImportState interface.
func (r *ResourceWithConfigureAndImportState) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}

// ImportState satisfies the resource.ResourceWithImportState interface.
func (r *ResourceWithConfigureAndImportState) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.ImportStateMethod == nil {
		return
	}

	r.ImportStateMethod(ctx, req, resp)
}
