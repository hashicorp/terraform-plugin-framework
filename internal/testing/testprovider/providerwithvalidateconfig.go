// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ provider.Provider = &ProviderWithValidateConfig{}
var _ provider.ProviderWithValidateConfig = &ProviderWithValidateConfig{}

// Declarative provider.ProviderWithValidateConfig for unit testing.
type ProviderWithValidateConfig struct {
	*Provider

	// ProviderWithValidateConfig interface methods
	ValidateConfigMethod func(context.Context, provider.ValidateConfigRequest, *provider.ValidateConfigResponse)
}

// GetMetaSchema satisfies the provider.ProviderWithValidateConfig interface.
func (p *ProviderWithValidateConfig) ValidateConfig(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	if p.ValidateConfigMethod == nil {
		return
	}

	p.ValidateConfigMethod(ctx, req, resp)
}
