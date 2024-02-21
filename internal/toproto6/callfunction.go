// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// CallFunctionResponse returns the *tfprotov6.CallFunctionResponse
// equivalent of a *fwserver.CallFunctionResponse.
func CallFunctionResponse(ctx context.Context, fw *fwserver.CallFunctionResponse) *tfprotov6.CallFunctionResponse {
	if fw == nil {
		return nil
	}

	proto := &tfprotov6.CallFunctionResponse{
		Diagnostics: Diagnostics(ctx, fw.Diagnostics),
	}

	result, diags := FunctionResultData(ctx, fw.Result)

	proto.Diagnostics = append(proto.Diagnostics, Diagnostics(ctx, diags)...)
	proto.Result = result

	return proto
}
