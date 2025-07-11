// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// InvokeActionRequest is the framework server request for the InvokeAction RPC.
type InvokeActionRequest struct {
	ActionSchema fwschema.Schema
	Config       *tfsdk.Config
}

// InvokeActionEventsStream is the framework server stream for the InvokeAction RPC.
type InvokeActionResponse struct {
	Diagnostics diag.Diagnostics
}

// InvokeAction implements the framework server InvokeAction RPC.
func (s *Server) InvokeAction(ctx context.Context, req *InvokeActionRequest, resp *InvokeActionResponse) {
	// TODO:Actions: Implementation coming soon...
	resp.Diagnostics.AddError(
		"InvokeAction Not Implemented",
		"InvokeAction has not yet been implemented in terraform-plugin-framework.",
	)
}
