// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.EphemeralResource = &EphemeralResource{}

// Declarative ephemeral.EphemeralResource for unit testing.
type EphemeralResource struct {
	// EphemeralResource interface methods
	MetadataMethod func(context.Context, ephemeral.MetadataRequest, *ephemeral.MetadataResponse)
	SchemaMethod   func(context.Context, ephemeral.SchemaRequest, *ephemeral.SchemaResponse)
	OpenMethod     func(context.Context, ephemeral.OpenRequest, *ephemeral.OpenResponse)
}

// Metadata satisfies the ephemeral.EphemeralResource interface.
func (r *EphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

// Schema satisfies the ephemeral.EphemeralResource interface.
func (r *EphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	if r.SchemaMethod == nil {
		return
	}

	r.SchemaMethod(ctx, req, resp)
}

// Open satisfies the ephemeral.EphemeralResource interface.
func (r *EphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	if r.OpenMethod == nil {
		return
	}

	r.OpenMethod(ctx, req, resp)
}
