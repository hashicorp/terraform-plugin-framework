// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Bool = &Bool{}

// Declarative planmodifier.Bool for unit testing.
type Bool struct {
	// Bool interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyBoolMethod      func(context.Context, planmodifier.BoolRequest, *planmodifier.BoolResponse)
}

// Description satisfies the planmodifier.Bool interface.
func (v Bool) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Bool interface.
func (v Bool) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Bool interface.
func (v Bool) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if v.PlanModifyBoolMethod == nil {
		return
	}

	v.PlanModifyBoolMethod(ctx, req, resp)
}
