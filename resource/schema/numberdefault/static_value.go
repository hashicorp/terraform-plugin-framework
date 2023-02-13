package numberdefault

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static number value default handler.
//
// Use StaticValue if a static default value for a number should be set.
func StaticValue(defaultVal types.Number) defaults.Number {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a number attribute.
type staticValueDefault struct {
	defaultVal types.Number
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %v", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%v`", d.defaultVal)
}

// DefaultNumber implements the static default value logic.
func (d staticValueDefault) DefaultNumber(ctx context.Context, req defaults.NumberRequest, resp *defaults.NumberResponse) {
	resp.PlanValue = d.defaultVal
}
