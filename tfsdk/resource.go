package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// A ResourceType is a type of resource. For each type of resource this provider
// supports, it should define a type implementing ResourceType and return an
// instance of it in the map returned by Provider.GetResources.
type ResourceType interface {
	// GetSchema returns the schema for this resource.
	GetSchema(context.Context) (Schema, diag.Diagnostics)

	// NewResource instantiates a new Resource of this ResourceType.
	NewResource(context.Context, Provider) (Resource, diag.Diagnostics)
}

// Resource represents a resource instance. This is the core interface that all
// resources must implement.
type Resource interface {
	// Create is called when the provider must create a new resource. Config
	// and planned state values should be read from the
	// CreateResourceRequest and new state values set on the
	// CreateResourceResponse.
	Create(context.Context, CreateResourceRequest, *CreateResourceResponse)

	// Read is called when the provider must read resource values in order
	// to update state. Planned state values should be read from the
	// ReadResourceRequest and new state values set on the
	// ReadResourceResponse.
	Read(context.Context, ReadResourceRequest, *ReadResourceResponse)

	// Update is called to update the state of the resource. Config, planned
	// state, and prior state values should be read from the
	// UpdateResourceRequest and new state values set on the
	// UpdateResourceResponse.
	Update(context.Context, UpdateResourceRequest, *UpdateResourceResponse)

	// Delete is called when the provider must delete the resource. Config
	// values may be read from the DeleteResourceRequest.
	Delete(context.Context, DeleteResourceRequest, *DeleteResourceResponse)

	// ImportState is called when the provider must import the resource.
	//
	// If import is not supported, it is recommended to use the
	// ResourceImportStateNotImplemented() call in this method.
	//
	// If setting an attribute with the import identifier, it is recommended
	// to use the ResourceImportStatePassthroughID() call in this method.
	ImportState(context.Context, ImportResourceStateRequest, *ImportResourceStateResponse)
}

// ResourceWithModifyPlan represents a resource instance with a ModifyPlan
// function.
type ResourceWithModifyPlan interface {
	Resource

	// ModifyPlan is called when the provider has an opportunity to modify
	// the plan: once during the plan phase when Terraform is determining
	// the diff that should be shown to the user for approval, and once
	// during the apply phase with any unknown values from configuration
	// filled in with their final values.
	//
	// The planned new state is represented by
	// ModifyResourcePlanResponse.Plan. It must meet the following
	// constraints:
	// 1. Any non-Computed attribute set in config must preserve the exact
	// config value or return the corresponding attribute value from the
	// prior state (ModifyResourcePlanRequest.State).
	// 2. Any attribute with a known value must not have its value changed
	// in subsequent calls to ModifyPlan or Create/Read/Update.
	// 3. Any attribute with an unknown value may either remain unknown
	// or take on any value of the expected type.
	//
	// Any errors will prevent further resource-level plan modifications.
	ModifyPlan(context.Context, ModifyResourcePlanRequest, *ModifyResourcePlanResponse)
}
