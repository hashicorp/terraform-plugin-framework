// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package setdefault

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static set value default handler.
//
// Use StaticValue if a static default value for a set should be set.
func StaticValue(defaultVal types.Set) defaults.Set {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a set attribute.
type staticValueDefault struct {
	defaultVal types.Set
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %v", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%v`", d.defaultVal)
}

// DefaultSet implements the static default value logic.
func (d staticValueDefault) DefaultSet(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
	resp.PlanValue = d.defaultVal
}
