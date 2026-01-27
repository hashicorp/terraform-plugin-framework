// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"bytes"
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// readStateBytesErrorDiagnostics returns a value suitable for
// [ReadStateBytesServerStream.Chunks]. It yields a single result that contains
// the given error diagnostics.
func readStateBytesErrorDiagnostics(ctx context.Context, diags diag.Diagnostics) (*tfprotov6.ReadStateBytesStream, error) {
	return &tfprotov6.ReadStateBytesStream{
		Chunks: func(push func(chunk tfprotov6.ReadStateByteChunk) bool) {
			push(tfprotov6.ReadStateByteChunk{
				Diagnostics: toproto6.Diagnostics(ctx, diags), // TODO : Think about how we handle diags
			})
		},
	}, nil
}

// ReadStateBytes satisfies the tfprotov6.ProviderServer interface.
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
				push(tfprotov6.ReadStateByteChunk{
					Diagnostics: toproto6.Diagnostics(ctx, fwResp.Diagnostics),
				})
				return
			}

			// Use chunk size from server capabilities, default to 8MB if not set
			chunkSize := fwResp.ServerCapabilities.ChunkSize
			if chunkSize == 0 {
				chunkSize = 8 << 20 // 8 MB default
			}

			reader := bytes.NewReader(fwResp.Bytes)
			totalLength := reader.Size()
			rangeStart := 0

			for {
				readBytes := make([]byte, chunkSize)
				byteCount, err := reader.Read(readBytes)

				if byteCount == 0 {
					break
				}

				chunk := tfprotov6.ReadStateByteChunk{
					StateByteChunk: tfprotov6.StateByteChunk{
						Bytes:       readBytes[:byteCount],
						TotalLength: totalLength,
						Range: tfprotov6.StateByteRange{
							Start: int64(rangeStart),
							End:   int64(rangeStart + byteCount),
						},
					},
				}

				// Include diagnostics only on first chunk
				if rangeStart == 0 {
					chunk.Diagnostics = toproto6.Diagnostics(ctx, fwResp.Diagnostics)
				}

				if !push(chunk) {
					return
				}

				rangeStart += byteCount

				// Handle read errors
				if err != nil {
					break
				}
			}
		},
	}

	return protoStream, nil
}
