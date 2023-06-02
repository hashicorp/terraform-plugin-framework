// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ provider.Provider = &ProviderWithConfigValidators{}
var _ provider.ProviderWithConfigValidators = &ProviderWithConfigValidators{}

// Declarative provider.ProviderWithConfigValidators for unit testing.
type ProviderWithConfigValidators struct {
	*Provider

	// ProviderWithConfigValidators interface methods
	ConfigValidatorsMethod func(context.Context) []provider.ConfigValidator
}

// GetMetaSchema satisfies the provider.ProviderWithConfigValidators interface.
func (p *ProviderWithConfigValidators) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	if p.ConfigValidatorsMethod == nil {
		return nil
	}

	return p.ConfigValidatorsMethod(ctx)
}
