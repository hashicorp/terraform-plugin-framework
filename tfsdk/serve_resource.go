package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func serveGetResourceType(ctx context.Context, provider Provider, typ string) (ResourceType, diag.Diagnostics) {
	resourceTypes, diags := provider.GetResources(ctx)
	if diags.HasError() {
		return nil, diags
	}
	resourceType, ok := resourceTypes[typ]
	if !ok {
		diags.AddError(
			"Resource not found",
			fmt.Sprintf("No resource named %q is configured on the provider", typ),
		)
		return nil, diags
	}
	return resourceType, diags
}

func serveGetResourceSchema(ctx context.Context, provider Provider, typ string) (Schema, diag.Diagnostics) {
	resourceType, diags := serveGetResourceType(ctx, provider, typ)
	if diags.HasError() {
		return Schema{}, diags
	}
	schema, ds := resourceType.GetSchema(ctx)
	diags.Append(ds...)
	if diags.HasError() {
		return Schema{}, diags
	}
	return schema, diags
}

func serveGetResourceInstance(ctx context.Context, provider Provider, typ string) (Resource, diag.Diagnostics) {
	resourceType, diags := serveGetResourceType(ctx, provider, typ)
	if diags.HasError() {
		return nil, diags
	}
	resource, ds := resourceType.NewResource(ctx, provider)
	diags.Append(ds...)
	if diags.HasError() {
		return nil, diags
	}
	return resource, diags
}
