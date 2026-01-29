// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func (s *Server) WriteStateBytes(ctx context.Context, proto6Req *tfprotov6.WriteStateBytesStream) (*tfprotov6.WriteStateBytesResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	var typeName string
	var stateID string

	var stateBuffer bytes.Buffer

	// TODO: Is order guaranteed here? Maybe check proto documentation for client side streaming on order consistency?
	// TODO: do anything with range? Probably not?
	for chunk, diags := range proto6Req.Chunks {
		// GRPC errors, client close, invalid data from client, etc.
		if len(diags) > 0 {
			return &tfprotov6.WriteStateBytesResponse{
				Diagnostics: diags,
			}, nil
		}

		if chunk.Meta != nil {
			typeName = chunk.Meta.TypeName
			stateID = chunk.Meta.StateID
		}

		_, err := stateBuffer.Write(chunk.Bytes)
		if err != nil {
			// TODO: can this ever happen?
			return &tfprotov6.WriteStateBytesResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Error writing state",
						Detail:   fmt.Sprintf("%s", err.Error()), // todo: cleanup
					},
				},
			}, nil
		}
	}

	if stateBuffer.Len() == 0 {
		// TODO: this shouldn't happen, probably error
	}

	if stateID == "" {
		// todo: error
	}

	// TODO: validate that the total byte size recieved is equal to chunk.TotalLength

	fwResp := &fwserver.WriteStateBytesResponse{}

	stateStore, diags := s.FrameworkServer.StateStore(ctx, typeName)

	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return &tfprotov6.WriteStateBytesResponse{
			Diagnostics: toproto6.Diagnostics(ctx, diags),
		}, nil
	}

	fwReq := &fwserver.WriteStateBytesRequest{
		// Framework
		StateStore: stateStore,

		// Provider dev
		StateID: stateID,
		Data:    stateBuffer.Bytes(),
	}

	s.FrameworkServer.WriteStateBytes(ctx, fwReq, fwResp)

	return &tfprotov6.WriteStateBytesResponse{
		Diagnostics: toproto6.Diagnostics(ctx, fwResp.Diagnostics),
	}, nil
}
