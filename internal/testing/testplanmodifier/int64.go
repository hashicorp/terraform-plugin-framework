// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Int64 = &Int64{}

// Declarative planmodifier.Int64 for unit testing.
type Int64 struct {
	// Int64 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyInt64Method     func(context.Context, planmodifier.Int64Request, *planmodifier.Int64Response)
}

// Description satisfies the planmodifier.Int64 interface.
func (v Int64) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Int64 interface.
func (v Int64) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Int64 interface.
func (v Int64) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if v.PlanModifyInt64Method == nil {
		return
	}

	v.PlanModifyInt64Method(ctx, req, resp)
}
