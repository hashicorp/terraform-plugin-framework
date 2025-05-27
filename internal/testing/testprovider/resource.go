// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ list.ListResource = &ListResource{}

// Declarative list.ListResource for unit testing.
type ListResource struct {
	// ListResource interface methods
	MetadataMethod                 func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)
	ListResourceConfigSchemaMethod func(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	ListResourceMethod             func(context.Context, list.ListResourceRequest, *list.ListResourceResponse)
}

// Metadata satisfies the list.ListResource interface.
func (r *ListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

// ListResourceConfigSchema satisfies the list.ListResource interface.
func (r *ListResource) ListResourceConfigSchema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	if r.ListResourceConfigSchemaMethod == nil {
		return
	}

	r.ListResourceConfigSchemaMethod(ctx, req, resp)
}

// ListResource satisfies the list.ListResource interface.
func (r *ListResource) ListResource(ctx context.Context, req list.ListResourceRequest, resp *list.ListResourceResponse) {
	if r.ListResourceMethod == nil {
		return
	}
	r.ListResourceMethod(ctx, req, resp)
}
