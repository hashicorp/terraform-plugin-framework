package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.ResourceType = &ResourceType{}

// Declarative tfsdk.ResourceType for unit testing.
type ResourceType struct {
	// ResourceType interface methods
	GetSchemaMethod   func(context.Context) (tfsdk.Schema, diag.Diagnostics)
	NewResourceMethod func(context.Context, tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics)
}

// GetSchema satisfies the tfsdk.ResourceType interface.
func (t *ResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if t.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return t.GetSchemaMethod(ctx)
}

// NewResource satisfies the tfsdk.ResourceType interface.
func (t *ResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	if t.NewResourceMethod == nil {
		return nil, nil
	}

	return t.NewResourceMethod(ctx, p)
}
