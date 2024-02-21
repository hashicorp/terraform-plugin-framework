// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
)

// CallFunctionRequest is the framework server request for the
// CallFunction RPC.
type CallFunctionRequest struct {
	Arguments          function.ArgumentsData
	Function           function.Function
	FunctionDefinition function.Definition
}

// CallFunctionResponse is the framework server response for the
// CallFunction RPC.
type CallFunctionResponse struct {
	Errors function.FunctionErrors
	Result function.ResultData
}

// CallFunction implements the framework server CallFunction RPC.
func (s *Server) CallFunction(ctx context.Context, req *CallFunctionRequest, resp *CallFunctionResponse) {
	var fe function.FunctionErrors

	if req == nil {
		return
	}

	resultData, err := req.FunctionDefinition.Return.NewResultData(ctx)

	fe.Append(err...)

	resp.Errors = fe

	if resp.Errors.HasError() {
		return
	}

	runReq := function.RunRequest{
		Arguments: req.Arguments,
	}
	runResp := function.RunResponse{
		Result: resultData,
	}

	logging.FrameworkTrace(ctx, "Calling provider defined Function Run")
	req.Function.Run(ctx, runReq, &runResp)
	logging.FrameworkTrace(ctx, "Called provider defined Function Run")

	resp.Errors = runResp.Errors
	resp.Result = runResp.Result
}
