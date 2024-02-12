// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

// CallFunctionResponse returns the *tfprotov6.CallFunctionResponse
// equivalent of a *fwserver.CallFunctionResponse.
func CallFunctionResponse(ctx context.Context, fw *fwserver.CallFunctionResponse) *tfprotov6.CallFunctionResponse {
	if fw == nil {
		return nil
	}

	funcErrs := fw.Error

	result, resultErrs := FunctionResultData(ctx, fw.Result)

	funcErrs.Append(resultErrs...)

	return &tfprotov6.CallFunctionResponse{
		Error:  FunctionError(ctx, funcErrs),
		Result: result,
	}
}
