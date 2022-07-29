package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// A ResourceType is a type of resource. For each type of resource this provider
// supports, it should define a type implementing ResourceType and return an
// instance of it in the map returned by Provider.GetResources.
type ResourceType interface {
	// GetSchema returns the schema for this resource.
	GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics)

	// NewResource instantiates a new Resource of this ResourceType.
	NewResource(context.Context, Provider) (resource.Resource, diag.Diagnostics)
}
