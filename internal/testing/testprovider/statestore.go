// Copyright IBM Corp. 2021, 2025
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
	MetadataMethod   func(context.Context, statestore.MetadataRequest, *statestore.MetadataResponse)
	SchemaMethod     func(context.Context, statestore.SchemaRequest, *statestore.SchemaResponse)
	InitializeMethod func(context.Context, statestore.InitializeRequest, *statestore.InitializeResponse)
	LockMethod       func(context.Context, statestore.LockRequest, *statestore.LockResponse)
	UnlockMethod     func(context.Context, statestore.UnlockRequest, *statestore.UnlockResponse)
	ReadMethod       func(context.Context, statestore.ReadRequest, *statestore.ReadResponse)
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

// Initialize satisfies the statestore.StateStore interface.
func (d *StateStore) Initialize(ctx context.Context, req statestore.InitializeRequest, resp *statestore.InitializeResponse) {
	if d.InitializeMethod == nil {
		return
	}

	d.InitializeMethod(ctx, req, resp)
}

// Lock satisfies the statestore.StateStore interface.
func (d *StateStore) Lock(ctx context.Context, req statestore.LockRequest, resp *statestore.LockResponse) {
	if d.LockMethod == nil {
		return
	}

	d.LockMethod(ctx, req, resp)
}

// Unlock satisfies the statestore.StateStore interface.
func (d *StateStore) Unlock(ctx context.Context, req statestore.UnlockRequest, resp *statestore.UnlockResponse) {
	if d.UnlockMethod == nil {
		return
	}

	d.UnlockMethod(ctx, req, resp)
}

// Read satisfies the statestore.StateStore interface.
func (d *StateStore) Read(ctx context.Context, req statestore.ReadRequest, resp *statestore.ReadResponse) {
	if d.ReadMethod == nil {
		return
	}

	d.ReadMethod(ctx, req, resp)
}
