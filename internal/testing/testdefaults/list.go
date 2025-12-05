// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.List = List{}

// Declarative defaults.List for unit testing.
type List struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.List interface methods
	DefaultListMethod func(context.Context, defaults.ListRequest, *defaults.ListResponse)
}

// Description satisfies the defaults.Describer interface.
func (v List) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v List) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultList satisfies the defaults.List interface.
func (v List) DefaultList(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
	if v.DefaultListMethod == nil {
		return
	}

	v.DefaultListMethod(ctx, req, resp)
}
