package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.AttributePlanModifier = &AttributePlanModifier{}

// Declarative tfsdk.AttributePlanModifier for unit testing.
type AttributePlanModifier struct {
	// AttributePlanModifier interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ModifyMethod              func(context.Context, tfsdk.ModifyAttributePlanRequest, *tfsdk.ModifyAttributePlanResponse)
}

// Description satisfies the tfsdk.AttributePlanModifier interface.
func (m *AttributePlanModifier) Description(ctx context.Context) string {
	if m.DescriptionMethod == nil {
		return ""
	}

	return m.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the tfsdk.AttributePlanModifier interface.
func (m *AttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	if m.MarkdownDescriptionMethod == nil {
		return ""
	}

	return m.MarkdownDescriptionMethod(ctx)
}

// Modify satisfies the tfsdk.AttributePlanModifier interface.
func (m *AttributePlanModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if m.ModifyMethod == nil {
		return
	}

	m.ModifyMethod(ctx, req, resp)
}
