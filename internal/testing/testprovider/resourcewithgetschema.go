package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ resource.Resource = &ResourceWithGetSchema{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchema{}

// Declarative resource.ResourceWithGetSchema for unit testing.
type ResourceWithGetSchema struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (tfsdk.Schema, diag.Diagnostics)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchema) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return r.GetSchemaMethod(ctx)
}
