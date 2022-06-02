package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.Resource = &Resource{}

// Declarative tfsdk.Resource for unit testing.
type Resource struct {
	// Resource interface methods
	CreateMethod func(context.Context, tfsdk.CreateResourceRequest, *tfsdk.CreateResourceResponse)
	DeleteMethod func(context.Context, tfsdk.DeleteResourceRequest, *tfsdk.DeleteResourceResponse)
	ReadMethod   func(context.Context, tfsdk.ReadResourceRequest, *tfsdk.ReadResourceResponse)
	UpdateMethod func(context.Context, tfsdk.UpdateResourceRequest, *tfsdk.UpdateResourceResponse)
}

// Create satisfies the tfsdk.Resource interface.
func (r *Resource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if r.CreateMethod == nil {
		return
	}

	r.CreateMethod(ctx, req, resp)
}

// Delete satisfies the tfsdk.Resource interface.
func (r *Resource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	if r.DeleteMethod == nil {
		return
	}

	r.DeleteMethod(ctx, req, resp)
}

// Read satisfies the tfsdk.Resource interface.
func (r *Resource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	if r.ReadMethod == nil {
		return
	}

	r.ReadMethod(ctx, req, resp)
}

// Update satisfies the tfsdk.Resource interface.
func (r *Resource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	if r.UpdateMethod == nil {
		return
	}

	r.UpdateMethod(ctx, req, resp)
}
