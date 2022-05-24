package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Provider = &ProviderWithProviderMeta{}
var _ tfsdk.ProviderWithProviderMeta = &ProviderWithProviderMeta{}

// Declarative tfsdk.ProviderWithProviderMeta for unit testing.
type ProviderWithProviderMeta struct {
	*Provider

	// ProviderWithProviderMeta interface methods
	GetMetaSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)
}

// GetMetaSchema satisfies the tfsdk.ProviderWithProviderMeta interface.
func (p *ProviderWithProviderMeta) GetMetaSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if p.GetMetaSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return p.GetMetaSchemaMethod(ctx)
}
