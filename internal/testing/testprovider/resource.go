// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &Resource{}

// Declarative resource.Resource for unit testing.
type Resource struct {
	// Resource interface methods
	MetadataMethod func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)
	SchemaMethod   func(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	CreateMethod   func(context.Context, resource.CreateRequest, *resource.CreateResponse)
	DeleteMethod   func(context.Context, resource.DeleteRequest, *resource.DeleteResponse)
	ReadMethod     func(context.Context, resource.ReadRequest, *resource.ReadResponse)
	UpdateMethod   func(context.Context, resource.UpdateRequest, *resource.UpdateResponse)
}

// Metadata satisfies the resource.Resource interface.
func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}

// Schema satisfies the resource.Resource interface.
func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	if r.SchemaMethod == nil {
		return
	}

	r.SchemaMethod(ctx, req, resp)
}

// Create satisfies the resource.Resource interface.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.CreateMethod == nil {
		return
	}

	r.CreateMethod(ctx, req, resp)
}

// Delete satisfies the resource.Resource interface.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.DeleteMethod == nil {
		return
	}

	r.DeleteMethod(ctx, req, resp)
}

// Read satisfies the resource.Resource interface.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.ReadMethod == nil {
		return
	}

	r.ReadMethod(ctx, req, resp)
}

// Update satisfies the resource.Resource interface.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.UpdateMethod == nil {
		return
	}

	r.UpdateMethod(ctx, req, resp)
}
