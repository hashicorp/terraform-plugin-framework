// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package int32default

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticInt32 returns a static int32 value default handler.
//
// Use StaticInt32 if a static default value for a int32 should be set.
func StaticInt32(defaultVal int32) defaults.Int32 {
	return staticInt32Default{
		defaultVal: defaultVal,
	}
}

// staticInt32Default is static value default handler that
// sets a value on an int32 attribute.
type staticInt32Default struct {
	defaultVal int32
}

// Description returns a human-readable description of the default value handler.
func (d staticInt32Default) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %d", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticInt32Default) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%d`", d.defaultVal)
}

// DefaultInt32 implements the static default value logic.
func (d staticInt32Default) DefaultInt32(_ context.Context, req defaults.Int32Request, resp *defaults.Int32Response) {
	resp.PlanValue = types.Int32Value(d.defaultVal)
}
