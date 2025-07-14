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

	// TODO:Actions: Add linked resources once lifecycle/linked actions are implemented
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

	// TODO:Actions: Add linked resources once lifecycle/linked actions are implemented
}
