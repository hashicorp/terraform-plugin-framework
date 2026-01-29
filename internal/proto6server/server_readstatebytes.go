// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"bytes"
	"context"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func readStateBytesErrorDiagnostics(ctx context.Context, diags diag.Diagnostics) (*tfprotov6.ReadStateBytesStream, error) {
	return &tfprotov6.ReadStateBytesStream{
		Chunks: func(push func(chunk tfprotov6.ReadStateByteChunk) bool) {
			push(tfprotov6.ReadStateByteChunk{
				Diagnostics: toproto6.Diagnostics(ctx, diags),
			})
		},
	}, nil
}

func (s *Server) ReadStateBytes(ctx context.Context, proto6Req *tfprotov6.ReadStateBytesRequest) (*tfprotov6.ReadStateBytesStream, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.ReadStateBytesResponse{}

	statestore, diags := s.FrameworkServer.StateStore(ctx, proto6Req.TypeName)
	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return readStateBytesErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	statestoreSchema, diags := s.FrameworkServer.StateStoreSchema(ctx, proto6Req.TypeName)
	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return readStateBytesErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	fwReq, diags := fromproto6.ReadStateBytesRequest(ctx, proto6Req, statestore, statestoreSchema)
	fwResp.Diagnostics.Append(diags...)

	if fwResp.Diagnostics.HasError() {
		return readStateBytesErrorDiagnostics(ctx, fwResp.Diagnostics)
	}

	protoStream := &tfprotov6.ReadStateBytesStream{
		Chunks: func(push func(tfprotov6.ReadStateByteChunk) bool) {
			s.FrameworkServer.ReadStateBytes(ctx, fwReq, fwResp)

			if fwResp.Diagnostics.HasError() {
				if !push(tfprotov6.ReadStateByteChunk{
					Diagnostics: toproto6.Diagnostics(ctx, fwResp.Diagnostics),
				}) {
					return
				}
				return
			}

			chunkSize := 8 << 20 // default 8 MB
			if &fwResp.ServerCapabilities != nil && fwResp.ServerCapabilities.ChunkSize > 0 {
				chunkSize = int(fwResp.ServerCapabilities.ChunkSize)
			}

			reader := bytes.NewReader(fwResp.Bytes)
			totalLength := reader.Size()
			var rangeStart int64 = 0

			for {
				buf := make([]byte, chunkSize)
				byteCount, err := reader.Read(buf)

				if err != nil && err != io.EOF {
					chunk := tfprotov6.ReadStateByteChunk{
						Diagnostics: toproto6.Diagnostics(ctx, fwResp.Diagnostics),
					}
					if !push(chunk) {
						return
					}
					return
				}

				if byteCount == 0 {
					return
				}

				chunk := tfprotov6.ReadStateByteChunk{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       buf[:byteCount],
						TotalLength: totalLength,
						Range: tfprotov6.StateByteRange{
							Start: rangeStart,
							End:   rangeStart + int64(byteCount),
						},
					},
				}

				if rangeStart == 0 {
					chunk.Diagnostics = toproto6.Diagnostics(ctx, fwResp.Diagnostics)
				}

				if !push(chunk) {
					return
				}

				rangeStart += int64(byteCount)

				if err == io.EOF {
					return
				}
			}
		},
	}

	return protoStream, nil
}
