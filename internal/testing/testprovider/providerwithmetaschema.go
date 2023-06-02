// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ provider.Provider = &ProviderWithMetaSchema{}
var _ provider.ProviderWithMetaSchema = &ProviderWithMetaSchema{}

// Declarative provider.ProviderWithMetaSchema for unit testing.
type ProviderWithMetaSchema struct {
	*Provider

	// ProviderWithMetaSchema interface methods
	MetaSchemaMethod func(context.Context, provider.MetaSchemaRequest, *provider.MetaSchemaResponse)
}

// MetaSchema satisfies the provider.ProviderWithMetaSchema interface.
func (p *ProviderWithMetaSchema) MetaSchema(ctx context.Context, req provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
	if p.MetaSchemaMethod == nil {
		return
	}

	p.MetaSchemaMethod(ctx, req, resp)
}
