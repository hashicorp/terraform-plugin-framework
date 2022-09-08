package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithTypeName{}
var _ resource.ResourceWithTypeName = &ResourceWithTypeName{}

// Declarative resource.ResourceWithTypeName for unit testing.
type ResourceWithTypeName struct {
	*Resource

	// ResourceWithTypeName interface methods
	TypeNameMethod func(context.Context, resource.TypeNameRequest, *resource.TypeNameResponse)
}

// TypeName satisfies the resource.ResourceWithTypeName interface.
func (r *ResourceWithTypeName) TypeName(ctx context.Context, req resource.TypeNameRequest, resp *resource.TypeNameResponse) {
	if r.TypeNameMethod == nil {
		return
	}

	r.TypeNameMethod(ctx, req, resp)
}
