// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ provider.ConfigValidator = &ProviderConfigValidator{}

// Declarative provider.ConfigValidator for unit testing.
type ProviderConfigValidator struct {
	// ConfigValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateProviderMethod    func(context.Context, provider.ValidateConfigRequest, *provider.ValidateConfigResponse)
}

// Description satisfies the provider.ConfigValidator interface.
func (v *ProviderConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the provider.ConfigValidator interface.
func (v *ProviderConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the provider.ConfigValidator interface.
func (v *ProviderConfigValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	if v.ValidateProviderMethod == nil {
		return
	}

	v.ValidateProviderMethod(ctx, req, resp)
}
