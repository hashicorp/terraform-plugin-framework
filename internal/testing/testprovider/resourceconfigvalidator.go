// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ConfigValidator = &ResourceConfigValidator{}

// Declarative resource.ConfigValidator for unit testing.
type ResourceConfigValidator struct {
	// ResourceConfigValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateResourceMethod    func(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse)
}

// Description satisfies the resource.ConfigValidator interface.
func (v *ResourceConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the resource.ConfigValidator interface.
func (v *ResourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the resource.ConfigValidator interface.
func (v *ResourceConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if v.ValidateResourceMethod == nil {
		return
	}

	v.ValidateResourceMethod(ctx, req, resp)
}
