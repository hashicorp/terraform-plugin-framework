package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.ProviderConfigValidator = &ProviderConfigValidator{}

// Declarative tfsdk.ProviderConfigValidator for unit testing.
type ProviderConfigValidator struct {
	// ProviderConfigValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateMethod            func(context.Context, tfsdk.ValidateProviderConfigRequest, *tfsdk.ValidateProviderConfigResponse)
}

// Description satisfies the tfsdk.ProviderConfigValidator interface.
func (v *ProviderConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the tfsdk.ProviderConfigValidator interface.
func (v *ProviderConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the tfsdk.ProviderConfigValidator interface.
func (v *ProviderConfigValidator) Validate(ctx context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
