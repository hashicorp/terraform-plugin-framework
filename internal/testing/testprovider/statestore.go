// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

var _ statestore.StateStore = &StateStore{}

// Declarative statestore.StateStore for unit testing.
type StateStore struct {
	// StateStore interface methods
	MetadataMethod func(context.Context, statestore.MetadataRequest, *statestore.MetadataResponse)
	SchemaMethod   func(context.Context, statestore.SchemaRequest, *statestore.SchemaResponse)
}

// Metadata satisfies the statestore.StateStore interface.
func (d *StateStore) Metadata(ctx context.Context, req statestore.MetadataRequest, resp *statestore.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}

// Schema satisfies the statestore.StateStore interface.
func (d *StateStore) Schema(ctx context.Context, req statestore.SchemaRequest, resp *statestore.SchemaResponse) {
	if d.SchemaMethod == nil {
		return
	}

	d.SchemaMethod(ctx, req, resp)
}
