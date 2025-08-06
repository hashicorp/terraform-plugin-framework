// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// InvokeRequest represents a request for the provider to invoke the action and update
// the requested action's linked resources.
type InvokeRequest struct {
	// Config is the configuration the user supplied for the action.
	Config tfsdk.Config

	// LinkedResources contains the data of the managed resource types that are linked to this action.
	//
	//   - If the action schema type is Unlinked, this field will be empty.
	//   - If the action schema type is Lifecycle, this field will contain a single linked resource.
	//   - If the action schema type is Linked, this field will be one or more linked resources, which
	//     will be in the same order as the linked resource schemas are defined in the action schema.
	//
	// For Lifecycle actions, the provider may only change computed-only attributes of the linked resources.
	// For Linked actions, the provider may change any attributes of the linked resources.
	LinkedResources []InvokeRequestLinkedResource
}

// InvokeRequestLinkedResource represents linked resource data before the action is invoked.
type InvokeRequestLinkedResource struct {
	// Config is the configuration the user supplied for the linked resource.
	Config tfsdk.Config

	// State is the current state of the linked resource.
	State tfsdk.State

	// Identity is the planned identity of the linked resource. If the linked resource does not
	// support identity, this value will not be set.
	Identity *tfsdk.ResourceIdentity

	// Plan is the latest planned new state for the linked resource. This could
	// be the original plan, the result of the linked resource apply, or an invoke from a predecessor action.
	Plan tfsdk.Plan
}

// InvokeResponse represents a response to an InvokeRequest. An
// instance of this response struct is supplied as
// an argument to the action's Invoke function, in which the provider
// should set values on the InvokeResponse as appropriate.
type InvokeResponse struct {
	// Diagnostics report errors or warnings related to invoking the action or updating
	// the state of the requested action's linked resources. Returning an empty slice
	// indicates a successful invocation with no warnings or errors
	// generated.
	Diagnostics diag.Diagnostics

	// LinkedResources contains the provider modified data of the managed resource types that are linked to this action.
	//
	// For Lifecycle actions, the provider may only change computed-only attributes of the linked resources.
	// For Linked actions, the provider may change any attributes of the linked resources.
	LinkedResources []InvokeResponseLinkedResource

	// SendProgress will immediately send a progress update to Terraform core during action invocation.
	// This function is pre-populated by the framework and can be called multiple times while action logic is running.
	SendProgress func(event InvokeProgressEvent)
}

// InvokeResponseLinkedResource represents linked resource data that was changed during Invoke and returned.
type InvokeResponseLinkedResource struct {
	// State is the state of the linked resource following the Invoke operation.
	// This field is pre-populated from InvokeRequest.Plan and
	// should be set during the action's Invoke operation.
	State tfsdk.State

	// Identity is the identity of the linked resource following the Invoke operation.
	// This field is pre-populated from InvokeRequest.Identity and
	// should be set during the action's Invoke operation.
	Identity *tfsdk.ResourceIdentity

	// RequiresReplace indicates that the linked resource must be replaced as a result of an action invocation error.
	// This field can only be set to true if diagnostics are returned in [InvokeResponse], otherwise Framework will append
	// a provider implementation diagnostic to [InvokeResponse].
	RequiresReplace bool
}

// InvokeProgressEvent is the event returned to Terraform while an action is being invoked.
type InvokeProgressEvent struct {
	// Message is the string that will be presented to the practitioner either via the console
	// or an external system like HCP Terraform.
	Message string
}
