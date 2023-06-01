// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package float64default

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticFloat64 returns a static float64 value default handler.
//
// Use StaticFloat64 if a static default value for a float64 should be set.
func StaticFloat64(defaultVal float64) defaults.Float64 {
	return staticFloat64Default{
		defaultVal: defaultVal,
	}
}

// staticFloat64Default is static value default handler that
// sets a value on a float64 attribute.
type staticFloat64Default struct {
	defaultVal float64
}

// Description returns a human-readable description of the default value handler.
func (d staticFloat64Default) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %f", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticFloat64Default) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%f`", d.defaultVal)
}

// DefaultFloat64 implements the static default value logic.
func (d staticFloat64Default) DefaultFloat64(_ context.Context, req defaults.Float64Request, resp *defaults.Float64Response) {
	resp.PlanValue = types.Float64Value(d.defaultVal)
}
