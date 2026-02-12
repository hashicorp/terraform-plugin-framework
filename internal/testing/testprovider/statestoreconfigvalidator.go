// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

var _ statestore.ConfigValidator = &StateStoreConfigValidator{}

// Declarative statestore.ConfigValidator for unit testing.
type StateStoreConfigValidator struct {
	// StateStoreConfigValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateStateStoreMethod  func(context.Context, statestore.ValidateConfigRequest, *statestore.ValidateConfigResponse)
}

// Description satisfies the statestore.ConfigValidator interface.
func (v *StateStoreConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the statestore.ConfigValidator interface.
func (v *StateStoreConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the statestore.ConfigValidator interface.
func (v *StateStoreConfigValidator) ValidateStateStore(ctx context.Context, req statestore.ValidateConfigRequest, resp *statestore.ValidateConfigResponse) {
	if v.ValidateStateStoreMethod == nil {
		return
	}

	v.ValidateStateStoreMethod(ctx, req, resp)
}
