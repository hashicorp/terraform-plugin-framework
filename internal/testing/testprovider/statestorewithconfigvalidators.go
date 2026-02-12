// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

var _ statestore.StateStore = &StateStoreWithConfigValidators{}
var _ statestore.StateStoreWithConfigValidators = &StateStoreWithConfigValidators{}

// Declarative statestore.StateStoreWithConfigValidators for unit testing.
type StateStoreWithConfigValidators struct {
	*StateStore

	// StateStoreWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []statestore.ConfigValidator
}

// ConfigValidators satisfies the statestore.StateStoreWithConfigValidators interface.
func (p *StateStoreWithConfigValidators) ConfigValidators(ctx context.Context) []statestore.ConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
