// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
)

var _ list.ConfigValidator = &ListResourceConfigValidator{}

// Declarative list.ConfigValidator for unit testing.
type ListResourceConfigValidator struct {
	// ListResourceConfigValidator interface methods
	DescriptionMethod          func(context.Context) string
	MarkdownDescriptionMethod  func(context.Context) string
	ValidateListResourceMethod func(context.Context, list.ValidateConfigRequest, *list.ValidateConfigResponse)
}

// Description satisfies the list.ConfigValidator interface.
func (v *ListResourceConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the list.ConfigValidator interface.
func (v *ListResourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the list.ConfigValidator interface.
func (v *ListResourceConfigValidator) ValidateListResourceConfig(ctx context.Context, req list.ValidateConfigRequest, resp *list.ValidateConfigResponse) {
	if v.ValidateListResourceMethod == nil {
		return
	}

	v.ValidateListResourceMethod(ctx, req, resp)
}
