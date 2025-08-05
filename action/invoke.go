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

	LinkedResources []InvokeRequestLinkedResource
}

// TODO:Actions: docs, change name of this struct :?
type InvokeRequestLinkedResource struct {
	Config   tfsdk.Config
	State    tfsdk.State
	Identity *tfsdk.ResourceIdentity
	Plan     tfsdk.Plan
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

	LinkedResources []InvokeResponseLinkedResource

	// SendProgress will immediately send a progress update to Terraform core during action invocation.
	// This function is provided by the framework and can be called multiple times while action logic is running.
	//
	// TODO:Actions: More documentation about when you should use this / when you shouldn't
	SendProgress func(event InvokeProgressEvent)
}

// TODO:Actions: docs, change name of this struct :?
type InvokeResponseLinkedResource struct {
	State           tfsdk.State
	Identity        *tfsdk.ResourceIdentity
	RequiresReplace bool // TODO:Actions: Document that this can only be present when diagnostics exist
}

// InvokeProgressEvent is the event returned to Terraform while an action is being invoked.
type InvokeProgressEvent struct {
	// Message is the string that will be presented to the practitioner either via the console
	// or an external system like HCP Terraform.
	Message string
}
