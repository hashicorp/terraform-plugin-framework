package float64default

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static float64 value default handler.
//
// Use StaticValue if a static default value for a float64 should be set.
func StaticValue(defaultVal float64) defaults.Float64 {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a float64 attribute.
type staticValueDefault struct {
	defaultVal float64
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %f", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%f`", d.defaultVal)
}

// DefaultFloat64 implements the static default value logic.
func (d staticValueDefault) DefaultFloat64(_ context.Context, req defaults.Float64Request, resp *defaults.Float64Response) {
	resp.PlanValue = types.Float64Value(d.defaultVal)
}
