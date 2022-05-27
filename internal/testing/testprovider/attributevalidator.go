package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.AttributeValidator = &AttributeValidator{}

// Declarative tfsdk.AttributeValidator for unit testing.
type AttributeValidator struct {
	// AttributeValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateMethod            func(context.Context, tfsdk.ValidateAttributeRequest, *tfsdk.ValidateAttributeResponse)
}

// Description satisfies the tfsdk.AttributeValidator interface.
func (v *AttributeValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the tfsdk.AttributeValidator interface.
func (v *AttributeValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the tfsdk.AttributeValidator interface.
func (v *AttributeValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
