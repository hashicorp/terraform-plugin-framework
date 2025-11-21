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

// Configure satisfies the statestore.StateStore interface.
func (d *StateStore) Configure(ctx context.Context, request statestore.ConfigureStateStoreRequest, response *statestore.ConfigureStateStoreResponse) {
	if d.ConfigureMethod == nil {
		return
	}

	d.ConfigureMethod(ctx, request, response)
}

// Read satisfies the statestore.StateStore interface.
func (d *StateStore) Read(ctx context.Context, request statestore.ReadStateBytesRequest, response *statestore.ReadStateResponse) {
	if d.ReadMethod == nil {
		return
	}

	d.ReadMethod(ctx, request, response)
}

// Write satisfies the statestore.StateStore interface.
func (d *StateStore) Write(ctx context.Context, request statestore.WriteRequest, response *statestore.WriteResponse) {
	if d.WriteMethod == nil {
		return
	}

	d.WriteMethod(ctx, request, response)
}

// Lock satisfies the statestore.StateStore interface.
func (d *StateStore) Lock(ctx context.Context, request statestore.LockRequest, response *statestore.LockResponse) {
	if d.LockMethod == nil {
		return
	}

	d.LockMethod(ctx, request, response)
}

// Unlock satisfies the statestore.StateStore interface.
func (d *StateStore) Unlock(ctx context.Context, request statestore.UnlockStateRequest, response *statestore.UnlockStateResponse) {
	if d.UnlockMethod == nil {
		return
	}

	d.UnlockMethod(ctx, request, response)
}

// GetStates satisfies the statestore.StateStore interface.
func (d *StateStore) GetStates(ctx context.Context, request statestore.GetStatesRequest, response *statestore.GetStatesResponse) {
	if d.GetStatesMethod == nil {
		return
	}

	d.GetStatesMethod(ctx, request, response)
}

// DeleteState satisfies the statestore.StateStore interface.
func (d *StateStore) DeleteState(ctx context.Context, request statestore.DeleteStatesRequest, response *statestore.DeleteStatesResponse) {
	if d.DeleteStateMethod == nil {
		return
	}

	d.DeleteStateMethod(ctx, request, response)
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
