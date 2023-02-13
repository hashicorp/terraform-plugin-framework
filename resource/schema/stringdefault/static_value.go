package stringdefault

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static string value default handler.
//
// Use StaticValue if a static default value for a string should be set.
func StaticValue(defaultVal string) defaults.String {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a string attribute.
type staticValueDefault struct {
	defaultVal string
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %s", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%s`", d.defaultVal)
}

// DefaultString implements the static default value logic.
func (d staticValueDefault) DefaultString(_ context.Context, req defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.defaultVal)
}
