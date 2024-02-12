// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

// CallFunctionResponse returns the *tfprotov5.CallFunctionResponse
// equivalent of a *fwserver.CallFunctionResponse.
func CallFunctionResponse(ctx context.Context, fw *fwserver.CallFunctionResponse) *tfprotov5.CallFunctionResponse {
	if fw == nil {
		return nil
	}

	funcErrs := fw.Error

	result, resultErrs := FunctionResultData(ctx, fw.Result)

	funcErrs.Append(resultErrs...)

	return &tfprotov5.CallFunctionResponse{
		Error:  FunctionError(ctx, funcErrs),
		Result: result,
	}
}
