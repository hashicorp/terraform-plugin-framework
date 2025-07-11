// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// PlanActionRequest is the framework server request for the PlanAction RPC.
type PlanActionRequest struct {
	ActionSchema fwschema.Schema
	Config       *tfsdk.Config
}

// PlanActionResponse is the framework server response for the PlanAction RPC.
type PlanActionResponse struct {
	Diagnostics diag.Diagnostics
}

// PlanAction implements the framework server PlanAction RPC.
func (s *Server) PlanAction(ctx context.Context, req *PlanActionRequest, resp *PlanActionResponse) {
	// TODO:Actions: Implementation coming soon...
	resp.Diagnostics.AddError(
		"PlanAction Not Implemented",
		"PlanAction has not yet been implemented in terraform-plugin-framework.",
	)
}
