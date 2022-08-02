package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ provider.Provider = &ProviderWithMetaSchema{}
var _ provider.ProviderWithMetaSchema = &ProviderWithMetaSchema{}

// Declarative provider.ProviderWithMetaSchema for unit testing.
type ProviderWithMetaSchema struct {
	*Provider

	// ProviderWithMetaSchema interface methods
	GetMetaSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)
}

// GetMetaSchema satisfies the provider.ProviderWithMetaSchema interface.
func (p *ProviderWithMetaSchema) GetMetaSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if p.GetMetaSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return p.GetMetaSchemaMethod(ctx)
}
