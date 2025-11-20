// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

// ReadStateBytesRequest is the framework server request for the StateBytesResource RPC.
type ReadStateBytesRequest struct {
	StateStore statestore.StateStore
	StateId    string
}

// ReadStateBytesResponse is the framework server stream for the ReadStateBytes RPC.
type ReadStateBytesResponse struct {
	Bytes       []byte
	Diagnostics diag.Diagnostics
}

// ReadStateBytes implements the framework server ReadStateBytes RPC.
func (s *Server) ReadStateBytes(ctx context.Context, req *ReadStateBytesRequest, resp *ReadStateBytesResponse) {
	if req == nil {
		return
	}

	if statestoreWithConfigure, ok := req.StateStore.(statestore.StateStoreWithConfigure); ok {
		logging.FrameworkTrace(ctx, "StateStore implements StateStoreWithConfigure")

		configureReq := statestore.ConfigureStateStoreRequest{}
		configureResp := statestore.ConfigureStateStoreResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined StateStore Configure")
		statestoreWithConfigure.Configure(ctx, configureReq, &configureResp)
		logging.FrameworkTrace(ctx, "Called provider defined StateStore Configure")

		resp.Diagnostics = append(resp.Diagnostics, configureResp.Diagnostics...)
	}

	statestoreReq := statestore.ReadStateBytesRequest{StateId: req.StateId}
	statestoreResp := statestore.ReadStateResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined StateStore ReadStateBytes")
	req.StateStore.Read(ctx, statestoreReq, &statestoreResp)
	logging.FrameworkTrace(ctx, "Called provider defined StateStore ReadStateBytes")

	resp.Diagnostics = statestoreResp.Diagnostics
	resp.Bytes = statestoreResp.Bytes
}
