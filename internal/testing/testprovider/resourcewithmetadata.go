package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithMetadata{}
var _ resource.ResourceWithMetadata = &ResourceWithMetadata{}

// Declarative resource.ResourceWithMetadata for unit testing.
type ResourceWithMetadata struct {
	*Resource

	// ResourceWithMetadata interface methods
	MetadataMethod func(context.Context, resource.MetadataRequest, *resource.MetadataResponse)
}

// Metadata satisfies the resource.ResourceWithMetadata interface.
func (r *ResourceWithMetadata) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.MetadataMethod == nil {
		return
	}

	r.MetadataMethod(ctx, req, resp)
}
