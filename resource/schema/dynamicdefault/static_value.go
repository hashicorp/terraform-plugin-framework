// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamicdefault

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static dynamic value default handler.
//
// Use StaticValue if a static default value for a dynamic value should be set.
func StaticValue(defaultVal types.Dynamic) defaults.Dynamic {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a dynamic attribute.
type staticValueDefault struct {
	defaultVal types.Dynamic
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %s", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%s`", d.defaultVal)
}

// DefaultDynamic implements the static default value logic.
func (d staticValueDefault) DefaultDynamic(_ context.Context, req defaults.DynamicRequest, resp *defaults.DynamicResponse) {
	resp.PlanValue = d.defaultVal
}
