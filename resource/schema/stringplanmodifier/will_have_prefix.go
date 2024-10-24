package stringplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// TODO: docs
func WillHavePrefix(prefix string) planmodifier.String {
	return willHavePrefixModifier{
		prefix: prefix,
	}
}

type willHavePrefixModifier struct {
	prefix string
}

func (m willHavePrefixModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value will have the prefix %q once it becomes known", m.prefix)
}

func (m willHavePrefixModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value will have the prefix %q once it becomes known", m.prefix)
}

func (m willHavePrefixModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineWithPrefix(m.prefix)
}
