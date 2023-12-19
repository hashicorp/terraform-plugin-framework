// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// CallFunctionResponse returns the *tfprotov5.CallFunctionResponse
// equivalent of a *fwserver.CallFunctionResponse.
func CallFunctionResponse(ctx context.Context, fw *fwserver.CallFunctionResponse) *tfprotov5.CallFunctionResponse {
	if fw == nil {
		return nil
	}

	proto := &tfprotov5.CallFunctionResponse{
		Diagnostics: Diagnostics(ctx, fw.Diagnostics),
	}

	result, diags := FunctionResultData(ctx, fw.Result)

	proto.Diagnostics = append(proto.Diagnostics, Diagnostics(ctx, diags)...)
	proto.Result = result

	return proto
}
