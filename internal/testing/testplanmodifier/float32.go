// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Float32 = &Float32{}

// Declarative planmodifier.Float32 for unit testing.
type Float32 struct {
	// Float32 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyFloat32Method   func(context.Context, planmodifier.Float32Request, *planmodifier.Float32Response)
}

// Description satisfies the planmodifier.Float32 interface.
func (v Float32) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Float32 interface.
func (v Float32) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Float32 interface.
func (v Float32) PlanModifyFloat32(ctx context.Context, req planmodifier.Float32Request, resp *planmodifier.Float32Response) {
	if v.PlanModifyFloat32Method == nil {
		return
	}

	v.PlanModifyFloat32Method(ctx, req, resp)
}
