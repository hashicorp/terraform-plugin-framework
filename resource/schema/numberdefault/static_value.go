// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package numberdefault

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticBigFloat returns a static number value default handler.
//
// Use StaticBigFloat if a static default value for a number should be set.
func StaticBigFloat(defaultVal *big.Float) defaults.Number {
	return staticBigFloatDefault{
		defaultVal: defaultVal,
	}
}

// staticBigFloatDefault is static value default handler that
// sets a value on a number attribute.
type staticBigFloatDefault struct {
	defaultVal *big.Float
}

// Description returns a human-readable description of the default value handler.
func (d staticBigFloatDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %v", d.defaultVal)
}

// MarkdownDescription returns a markdown description of the default value handler.
func (d staticBigFloatDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to `%v`", d.defaultVal)
}

// DefaultNumber implements the static default value logic.
func (d staticBigFloatDefault) DefaultNumber(ctx context.Context, req defaults.NumberRequest, resp *defaults.NumberResponse) {
	resp.PlanValue = types.NumberValue(d.defaultVal)
}
