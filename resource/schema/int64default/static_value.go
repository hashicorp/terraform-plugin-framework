package int64default

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static int64 value default handler.
//
// Use StaticValue if a static default value for a int64 should be set.
func StaticValue(defaultVal int64) defaults.Int64 {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a int64 attribute.
type staticValueDefault struct {
	defaultVal int64
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %d", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%d`", d.defaultVal)
}

// DefaultInt64 implements the static default value logic.
func (d staticValueDefault) DefaultInt64(_ context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.defaultVal)
}
