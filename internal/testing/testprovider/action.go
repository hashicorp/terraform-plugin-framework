// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

var _ action.Action = &Action{}

// Declarative action.Action for unit testing.
type Action struct {
	// Action interface methods
	MetadataMethod func(context.Context, action.MetadataRequest, *action.MetadataResponse)
	SchemaMethod   func(context.Context, action.SchemaRequest, *action.SchemaResponse)
}

// Metadata satisfies the action.Action interface.
func (d *Action) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}

// Schema satisfies the action.Action interface.
func (d *Action) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	if d.SchemaMethod == nil {
		return
	}

	d.SchemaMethod(ctx, req, resp)
}
