// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Float32 = Float32{}

// Declarative defaults.Float32 for unit testing.
type Float32 struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Float32 interface methods
	DefaultFloat32Method func(context.Context, defaults.Float32Request, *defaults.Float32Response)
}

// Description satisfies the defaults.Describer interface.
func (v Float32) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Float32) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultFloat32 satisfies the defaults.Float32 interface.
func (v Float32) DefaultFloat32(ctx context.Context, req defaults.Float32Request, resp *defaults.Float32Response) {
	if v.DefaultFloat32Method == nil {
		return
	}

	v.DefaultFloat32Method(ctx, req, resp)
}
