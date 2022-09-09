package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ provider.ResourceType = &ResourceType{} //nolint:staticcheck // Internal implementation

// Declarative provider.ResourceType for unit testing.
type ResourceType struct {
	// ResourceType interface methods
	GetSchemaMethod   func(context.Context) (tfsdk.Schema, diag.Diagnostics)
	NewResourceMethod func(context.Context, provider.Provider) (resource.Resource, diag.Diagnostics)
}

// GetSchema satisfies the provider.ResourceType interface.
func (t *ResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if t.GetSchemaMethod == nil {
		return tfsdk.Schema{}, nil
	}

	return t.GetSchemaMethod(ctx)
}

// NewResource satisfies the provider.ResourceType interface.
func (t *ResourceType) NewResource(ctx context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	if t.NewResourceMethod == nil {
		return nil, nil
	}

	return t.NewResourceMethod(ctx, p)
}
