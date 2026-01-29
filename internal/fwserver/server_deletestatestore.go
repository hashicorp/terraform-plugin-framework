// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

// DeleteStatesRequest is the framework server request for the
// DeleteStates RPC.
type DeleteStatesRequest struct{}

// DeleteStatesResponse is the framework server response for the
// DeleteStates RPC.
type DeleteStatesResponse struct {
	States             []statestore.StateStore
	Diagnostics        diag.Diagnostics
	ServerCapabilities *ServerCapabilities
}

// DeleteStates implements the framework server DeleteStates RPC.
func (s *Server) DeleteStates(ctx context.Context, req *DeleteStatesRequest, resp *DeleteStatesResponse) {
	resp.ServerCapabilities = s.ServerCapabilities()

	stateStores, diags := s.StateStore(ctx, s.providerTypeName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.States = append(resp.States, stateStores)
}
