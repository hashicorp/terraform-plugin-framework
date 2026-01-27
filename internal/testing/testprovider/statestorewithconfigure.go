// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

var _ statestore.StateStore = &StateStoreWithConfigure{}
var _ statestore.StateStoreWithConfigure = &StateStoreWithConfigure{}

// Declarative statestore.StateStoreWithConfigure for unit testing.
type StateStoreWithConfigure struct {
	*StateStore

	// StateStoreWithConfigure interface methods
	ConfigureMethod func(context.Context, statestore.ConfigureRequest, *statestore.ConfigureResponse)
}

// Configure satisfies the statestore.StateStoreWithConfigure interface.
func (r *StateStoreWithConfigure) Configure(ctx context.Context, req statestore.ConfigureRequest, resp *statestore.ConfigureResponse) {
	if r.ConfigureMethod == nil {
		return
	}

	r.ConfigureMethod(ctx, req, resp)
}
