// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

func ReadStateByteChunkType(ctx context.Context, chunk *fwserver.ReadStateBytesResponse) tfprotov6.ReadStateByteChunk {
	return tfprotov6.ReadStateByteChunk{
		StateByteChunk: tfprotov6.StateByteChunk{
			Bytes: chunk.Bytes,
		},
	}
}
