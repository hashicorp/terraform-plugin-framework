// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var (
	_ provider.Provider              = &ProviderWithFunctions{}
	_ provider.ProviderWithFunctions = &ProviderWithFunctions{}
)

// Declarative provider.ProviderWithFunctions for unit testing.
type ProviderWithFunctions struct {
	*Provider

	// ProviderWithFunctions interface methods
	FunctionsMethod func(context.Context) []func() function.Function
}

// Functions satisfies the provider.ProviderWithFunctions interface.
func (p *ProviderWithFunctions) Functions(ctx context.Context) []func() function.Function {
	if p.FunctionsMethod == nil {
		return nil
	}

	return p.FunctionsMethod(ctx)
}
