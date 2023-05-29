// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapdefault

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static map value default handler.
//
// Use StaticValue if a static default value for a map should be set.
func StaticValue(defaultVal types.Map) defaults.Map {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a map attribute.
type staticValueDefault struct {
	defaultVal types.Map
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %v", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%v`", d.defaultVal)
}

// DefaultMap implements the static default value logic.
func (d staticValueDefault) DefaultMap(ctx context.Context, req defaults.MapRequest, resp *defaults.MapResponse) {
	resp.PlanValue = d.defaultVal
}
