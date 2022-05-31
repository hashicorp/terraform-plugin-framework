package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.ResourceConfigValidator = &ResourceConfigValidator{}

// Declarative tfsdk.ResourceConfigValidator for unit testing.
type ResourceConfigValidator struct {
	// ResourceConfigValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateMethod            func(context.Context, tfsdk.ValidateResourceConfigRequest, *tfsdk.ValidateResourceConfigResponse)
}

// Description satisfies the tfsdk.ResourceConfigValidator interface.
func (v *ResourceConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the tfsdk.ResourceConfigValidator interface.
func (v *ResourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the tfsdk.ResourceConfigValidator interface.
func (v *ResourceConfigValidator) Validate(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
