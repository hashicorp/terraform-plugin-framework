// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

type ReadStateBytesRequest struct {
	StateID    string
	StateStore statestore.StateStore
}

type ReadStateBytesResponse struct {
	Bytes       []byte
	Diagnostics diag.Diagnostics
}

// ReadStateBytes implements the framework server ReadStateBytes RPC.
func (s *Server) ReadStateBytes(ctx context.Context, req *ReadStateBytesRequest, resp *ReadStateBytesResponse) {
	if req == nil {
		return
	}

	if stateStoreWithConfigure, ok := req.StateStore.(statestore.StateStoreWithConfigure); ok {
		logging.FrameworkTrace(ctx, "StateStore implements StateStoreWithConfigure")

		configureReq := statestore.ConfigureRequest{
			StateStoreData: s.StateStoreConfigureData.StateStoreConfigureData,
		}
		configureResp := statestore.ConfigureResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined StateStore Configure")
		stateStoreWithConfigure.Configure(ctx, configureReq, &configureResp)
		logging.FrameworkTrace(ctx, "Called provider defined StateStore Configure")

		resp.Diagnostics.Append(configureResp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	readReq := statestore.ReadStateBytesRequest{
		StateID: req.StateID,
	}

	readResp := statestore.ReadStateBytesResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined StateStore ReadStateBytes")
	req.StateStore.Read(ctx, readReq, &readResp)
	logging.FrameworkTrace(ctx, "Called provider defined StateStore ReadStateBytes")

	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.Bytes = readResp.StateBytes
}
