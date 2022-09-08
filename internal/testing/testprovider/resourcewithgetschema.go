package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.Resource = &ResourceWithGetSchema{}
var _ resource.ResourceWithGetSchema = &ResourceWithGetSchema{}

// Declarative resource.ResourceWithGetSchema for unit testing.
type ResourceWithGetSchema struct {
	*Resource

	// ResourceWithGetSchema interface methods
	GetSchemaMethod func(context.Context) (fwschema.Schema, diag.Diagnostics)
}

// GetSchema satisfies the resource.ResourceWithGetSchema interface.
func (r *ResourceWithGetSchema) GetSchema(ctx context.Context) (fwschema.Schema, diag.Diagnostics) {
	if r.GetSchemaMethod == nil {
		return nil, nil
	}

	return r.GetSchemaMethod(ctx)
}
