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
	MetadataMethod    func(context.Context, statestore.MetadataRequest, *statestore.MetadataResponse)
	SchemaMethod      func(context.Context, statestore.SchemaRequest, *statestore.SchemaResponse)
	ConfigureMethod   func(context.Context, statestore.ConfigureStateStoreRequest, *statestore.ConfigureStateStoreResponse)
	ReadMethod        func(context.Context, statestore.ReadStateBytesRequest, *statestore.ReadStateResponse)
	WriteMethod       func(context.Context, statestore.WriteRequest, *statestore.WriteResponse)
	LockMethod        func(context.Context, statestore.LockRequest, *statestore.LockResponse)
	UnlockMethod      func(context.Context, statestore.UnlockStateRequest, *statestore.UnlockStateResponse)
	GetStatesMethod   func(context.Context, statestore.GetStatesRequest, *statestore.GetStatesResponse)
	DeleteStateMethod func(context.Context, statestore.DeleteStatesRequest, *statestore.DeleteStatesResponse)
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

// Configure satisfies the statestore.StateStore interface.
func (d *StateStore) Configure(ctx context.Context, req statestore.ConfigureStateStoreRequest, resp *statestore.ConfigureStateStoreResponse) {
	if d.ConfigureMethod == nil {
		return
	}

	d.ConfigureMethod(ctx, req, resp)
}

// Read satisfies the statestore.StateStore interface.
func (d *StateStore) Read(ctx context.Context, req statestore.ReadStateBytesRequest, resp *statestore.ReadStateResponse) {
	if d.ReadMethod == nil {
		return
	}

	d.ReadMethod(ctx, req, resp)
}

// Write satisfies the statestore.StateStore interface.
func (d *StateStore) Write(ctx context.Context, req statestore.WriteRequest, resp *statestore.WriteResponse) {
	if d.WriteMethod == nil {
		return
	}

	d.WriteMethod(ctx, req, resp)
}

// Lock satisfies the statestore.StateStore interface.
func (d *StateStore) Lock(ctx context.Context, req statestore.LockRequest, resp *statestore.LockResponse) {
	if d.LockMethod == nil {
		return
	}

	d.LockMethod(ctx, req, resp)
}

// Unlock satisfies the statestore.StateStore interface.
func (d *StateStore) Unlock(ctx context.Context, req statestore.UnlockStateRequest, resp *statestore.UnlockStateResponse) {
	if d.UnlockMethod == nil {
		return
	}

	d.UnlockMethod(ctx, req, resp)
}

// GetStates satisfies the statestore.StateStore interface.
func (d *StateStore) GetStates(ctx context.Context, req statestore.GetStatesRequest, resp *statestore.GetStatesResponse) {
	if d.GetStatesMethod == nil {
		return
	}

	d.GetStatesMethod(ctx, req, resp)
}

// DeleteState satisfies the statestore.StateStore interface.
func (d *StateStore) DeleteState(ctx context.Context, req statestore.DeleteStatesRequest, resp *statestore.DeleteStatesResponse) {
	if d.DeleteStateMethod == nil {
		return
	}

	d.DeleteStateMethod(ctx, req, resp)
}
