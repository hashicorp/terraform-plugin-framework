// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Int32 = &Int32{}

// Declarative planmodifier.Int32 for unit testing.
type Int32 struct {
	// Int32 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyInt32Method     func(context.Context, planmodifier.Int32Request, *planmodifier.Int32Response)
}

// Description satisfies the planmodifier.Int32 interface.
func (v Int32) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Int32 interface.
func (v Int32) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Int32 interface.
func (v Int32) PlanModifyInt32(ctx context.Context, req planmodifier.Int32Request, resp *planmodifier.Int32Response) {
	if v.PlanModifyInt32Method == nil {
		return
	}

	v.PlanModifyInt32Method(ctx, req, resp)
}
