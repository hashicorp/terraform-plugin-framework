package booldefault

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static boolean value default handler.
//
// Use StaticValue if a static default value for a boolean should be set.
func StaticValue(defaultVal bool) defaults.Bool {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a boolean attribute.
type staticValueDefault struct {
	defaultVal bool
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %t", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%t`", d.defaultVal)
}

// DefaultBool implements the static default value logic.
func (d staticValueDefault) DefaultBool(_ context.Context, req defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.defaultVal)
}
