// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

var _ statestore.StateStore = &StateStoreWithValidateConfig{}
var _ statestore.StateStoreWithValidateConfig = &StateStoreWithValidateConfig{}

// Declarative statestore.StateStoreWithValidateConfig for unit testing.
type StateStoreWithValidateConfig struct {
	*StateStore

	// StateStoreWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, statestore.ValidateConfigRequest, *statestore.ValidateConfigResponse)
}

// ValidateConfig satisfies the statestore.StateStoreWithValidateConfig interface.
func (p *StateStoreWithValidateConfig) ValidateConfig(ctx context.Context, req statestore.ValidateConfigRequest, resp *statestore.ValidateConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
