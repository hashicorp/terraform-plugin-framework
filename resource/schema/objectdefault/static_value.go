// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package objectdefault

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticValue returns a static object value default handler.
//
// Use StaticValue if a static default value for a object should be set.
func StaticValue(defaultVal types.Object) defaults.Object {
	return staticValueDefault{
		defaultVal: defaultVal,
	}
}

// staticValueDefault is static value default handler that
// sets a value on a object attribute.
type staticValueDefault struct {
	defaultVal types.Object
}

// Description returns a human-readable description of the default value handler.
func (d staticValueDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %v", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticValueDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%v`", d.defaultVal)
}

// DefaultObject implements the static default value logic.
func (d staticValueDefault) DefaultObject(ctx context.Context, req defaults.ObjectRequest, resp *defaults.ObjectResponse) {
	resp.PlanValue = d.defaultVal
}
