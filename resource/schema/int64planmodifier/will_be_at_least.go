package int64planmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// TODO: docs
func WillBeAtLeast(minVal int64) planmodifier.Int64 {
	return willBeAtLeastModifier{
		min: minVal,
	}
}

type willBeAtLeastModifier struct {
	min int64
}

func (m willBeAtLeastModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Promises the value will be at least %d once it becomes known", m.min)
}

func (m willBeAtLeastModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Promises the value will be at least %d once it becomes known", m.min)
}

func (m willBeAtLeastModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineWithLowerBound(m.min, true)
}
