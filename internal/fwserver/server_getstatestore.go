// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

// GetStatesRequest is the framework server request for the
// GetStates RPC.
type GetStatesRequest struct{}

// GetStatesResponse is the framework server response for the
// GetStates RPC.
type GetStatesResponse struct {
	States             []statestore.StateStore
	Diagnostics        diag.Diagnostics
	ServerCapabilities *ServerCapabilities
}

// GetStates implements the framework server GetStates RPC.
func (s *Server) GetStates(ctx context.Context, req *GetStatesRequest, resp *GetStatesResponse) {
	stateStores, diags := s.StateStore(ctx, s.providerTypeName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.States = append(resp.States, stateStores)
}
