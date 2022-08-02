package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &Resource{}

// Declarative resource.Resource for unit testing.
type Resource struct {
	// Resource interface methods
	CreateMethod func(context.Context, resource.CreateRequest, *resource.CreateResponse)
	DeleteMethod func(context.Context, resource.DeleteRequest, *resource.DeleteResponse)
	ReadMethod   func(context.Context, resource.ReadRequest, *resource.ReadResponse)
	UpdateMethod func(context.Context, resource.UpdateRequest, *resource.UpdateResponse)
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
