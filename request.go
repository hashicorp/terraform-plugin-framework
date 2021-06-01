package tfsdk

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ConfigureProviderRequest represents a request containing the values the user
// specified for the provider configuration block, along with other runtime
// information from Terraform or the Plugin SDK. An instance of this request
// struct is supplied as an argument to the provider's Configure function.
type ConfigureProviderRequest struct {
	// TerraformVersion is the version of Terraform executing the request.
	// This is supplied for logging, analytics, and User-Agent purposes
	// only. Providers should not try to gate provider behavior on
	// Terraform versions.
	TerraformVersion string

	// Config is the configuration the user supplied for the provider. This
	// information should usually be persisted to the underlying type
	// that's implementing the Provider interface, for use in later
	// resource CRUD operations.
	Config *tftypes.Value
}

// CreateResourceRequest represents a request for the provider to create a
// resource. An instance of this request struct is supplied as an argument to
// the resource's Create function.
type CreateResourceRequest struct {
	// Config is the configuration the user supplied for the resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	// TODO uncomment when implemented
	// Config Config

	// Plan is the planned state for the resource.
	// TODO uncomment when implemented
	// Plan Plan
}

// ReadResourceRequest represents a request for the provider to read a
// resource, i.e., update values in state according to the real state of the
// resource. An instance of this request struct is supplied as an argument to
// the resource's Read function.
type ReadResourceRequest struct {
	// State is the current state of the resource prior to the Read
	// operation.
	// TODO uncomment when implemented
	// State State
}

// UpdateResourceRequest represents a request for the provider to update a
// resource. An instance of this request struct is supplied as an argument to
// the resource's Update function.
type UpdateResourceRequest struct {
	// Config is the configuration the user supplied for the resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	// TODO uncomment when implemented
	// Config Config

	// Plan is the planned state for the resource.
	// TODO uncomment when implemented
	// Plan Plan

	// State is the current state of the resource prior to the Update
	// operation.
	// TODO uncomment when implemented
	// State State
}

// DeleteResourceRequest represents a request for the provider to delete a
// resource. An instance of this request struct is supplied as an argument to
// the resource's Delete function.
type DeleteResourceRequest struct {
	// Config is the configuration the user supplied for the resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	// TODO uncomment when implemented
	// Config Config
}
