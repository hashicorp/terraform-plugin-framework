package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// A ResourceType is a type of resource. For each type of resource this provider
// supports, it should instantiate a struct implementing ResourceType and return
// it as a value in the map returned by Provider.GetResources.
type ResourceType interface {
	// GetSchema returns the schema for this resource.
	GetSchema(context.Context) (*tfprotov6.Schema, []*tfprotov6.Diagnostic)

	// NewResource instantiates a new Resource of this ResourceType.
	NewResource(Provider) (Resource, []*tfprotov6.Diagnostic)
}

// Resource represents a resource instance. This is the core interface that all
// resources must implement.
type Resource interface {
	// Create is called when the provider must create a new resource. Config
	// and planned state values should be read from the
	// CreateResourceRequest and new state values set on the
	// CreateResourceResponse.
	Create(context.Context, *CreateResourceRequest, *CreateResourceResponse)

	// Read is called when the provider must read resource values in order
	// to update state. Planned state values should be read from the
	// ReadResourceRequest and new state values set on the
	// ReadResourceResponse.
	Read(context.Context, *ReadResourceRequest, *ReadResourceResponse)

	// Update is called to update the state of the resource. Config, planned
	// state, and prior state values should be read from the
	// UpdateResourceRequest and new state values set on the
	// UpdateResourceResponse.
	Update(context.Context, *UpdateResourceRequest, *UpdateResourceResponse)

	// Delete is called when the provider must delete the resource. Config
	// values may be read from the DeleteResourceRequest.
	Delete(context.Context, *DeleteResourceRequest, *DeleteResourceResponse)
}

// Importer is a resource that can be imported.
type Importer interface {
	// TODO create ImportResource{Request,Response} types
	Import(context.Context, string) ([]*tfprotov6.ImportedResource, []*tfprotov6.Diagnostic)
}

// Validater is a resource with a Validate method.
type Validater interface {
	// TODO create ValidateResource{Request,Response} types
	Validate(context.Context, tftypes.Value) []*tfprotov6.Diagnostic
}

// CustomPlanner is a resource with a ModifyPlan method.
type CustomPlanner interface {
	// TODO create PlanResource{Request,Response} types
	ModifyPlan(ctx context.Context, priorState, proposedState, config tftypes.Value) (tftypes.Value, []*tfprotov6.Diagnostic)
}
