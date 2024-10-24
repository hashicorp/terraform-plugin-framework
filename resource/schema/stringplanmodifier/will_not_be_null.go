package stringplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// TODO: docs
func WillNotBeNull() planmodifier.String {
	return willNotBeNullModifier{}
}

type willNotBeNullModifier struct{}

func (m willNotBeNullModifier) Description(_ context.Context) string {
	return "Promises the value will not be null once it becomes known"
}

func (m willNotBeNullModifier) MarkdownDescription(_ context.Context) string {
	return "Promises the value will not be null once it becomes known"
}

func (m willNotBeNullModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.PlanValue.RefineAsNotNull()
}
