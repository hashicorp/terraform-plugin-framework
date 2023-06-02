// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Number = &Number{}

// Declarative planmodifier.Number for unit testing.
type Number struct {
	// Number interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyNumberMethod    func(context.Context, planmodifier.NumberRequest, *planmodifier.NumberResponse)
}

// Description satisfies the planmodifier.Number interface.
func (v Number) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Number interface.
func (v Number) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Number interface.
func (v Number) PlanModifyNumber(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
	if v.PlanModifyNumberMethod == nil {
		return
	}

	v.PlanModifyNumberMethod(ctx, req, resp)
}
