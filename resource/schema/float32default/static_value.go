// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package float32default

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticFloat32 returns a static float32 value default handler.
//
// Use StaticFloat32 if a static default value for a float32 should be set.
func StaticFloat32(defaultVal float32) defaults.Float32 {
	return staticFloat32Default{
		defaultVal: defaultVal,
	}
}

// staticFloat32Default is static value default handler that
// sets a value on a float32 attribute.
type staticFloat32Default struct {
	defaultVal float32
}

// Description returns a human-readable description of the default value handler.
func (d staticFloat32Default) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %f", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticFloat32Default) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%f`", d.defaultVal)
}

// DefaultFloat32 implements the static default value logic.
func (d staticFloat32Default) DefaultFloat32(_ context.Context, req defaults.Float32Request, resp *defaults.Float32Response) {
	resp.PlanValue = types.Float32Value(d.defaultVal)
}
