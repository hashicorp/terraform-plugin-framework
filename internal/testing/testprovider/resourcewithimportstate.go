// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithImportState{}
var _ resource.ResourceWithImportState = &ResourceWithImportState{}

// Declarative resource.ResourceWithImportState for unit testing.
type ResourceWithImportState struct {
	*Resource

	// ResourceWithImportState interface methods
	ImportStateMethod func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)
}

// ImportState satisfies the resource.ResourceWithImportState interface.
func (p *ResourceWithImportState) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if p.ImportStateMethod == nil {
		return
	}

	p.ImportStateMethod(ctx, req, resp)
}
