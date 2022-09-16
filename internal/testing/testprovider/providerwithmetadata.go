package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

var _ provider.Provider = &ProviderWithMetadata{}
var _ provider.ProviderWithMetadata = &ProviderWithMetadata{}

// Declarative provider.ProviderWithMetadata for unit testing.
type ProviderWithMetadata struct {
	*Provider

	// ProviderWithMetadata interface methods
	MetadataMethod func(context.Context, provider.MetadataRequest, *provider.MetadataResponse)
}

// Metadata satisfies the provider.ProviderWithMetadata interface.
func (p *ProviderWithMetadata) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	if p.MetadataMethod == nil {
		return
	}

	p.MetadataMethod(ctx, req, resp)
}
