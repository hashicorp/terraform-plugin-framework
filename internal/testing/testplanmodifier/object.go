// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Object = &Object{}

// Declarative planmodifier.Object for unit testing.
type Object struct {
	// Object interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyObjectMethod    func(context.Context, planmodifier.ObjectRequest, *planmodifier.ObjectResponse)
}

// Description satisfies the planmodifier.Object interface.
func (v Object) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Object interface.
func (v Object) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Object interface.
func (v Object) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if v.PlanModifyObjectMethod == nil {
		return
	}

	v.PlanModifyObjectMethod(ctx, req, resp)
}
