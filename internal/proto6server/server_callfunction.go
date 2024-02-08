// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
)

// CallFunction satisfies the tfprotov6.ProviderServer interface.
func (s *Server) CallFunction(ctx context.Context, protoReq *tfprotov6.CallFunctionRequest) (*tfprotov6.CallFunctionResponse, error) {
	ctx = s.registerContext(ctx)
	ctx = logging.InitContext(ctx)

	fwResp := &fwserver.CallFunctionResponse{}

	function, err := s.FrameworkServer.Function(ctx, protoReq.Name)

	fwResp.Error = err

	if fwResp.Error != nil {
		//nolint:nilerr
		return toproto6.CallFunctionResponse(ctx, fwResp), nil
	}

	functionDefinition, diags := s.FrameworkServer.FunctionDefinition(ctx, protoReq.Name)

	for _, d := range diags {
		fwResp.Error = errors.Join(fwResp.Error, fmt.Errorf("%s: %s\n\n%s", d.Severity(), d.Summary(), d.Detail()))
	}

	if fwResp.Error != nil {
		//nolint:nilerr
		return toproto6.CallFunctionResponse(ctx, fwResp), nil
	}

	fwReq, diags := fromproto6.CallFunctionRequest(ctx, protoReq, function, functionDefinition)

	for _, d := range diags {
		fwResp.Error = errors.Join(fwResp.Error, fmt.Errorf("%s: %s\n\n%s", d.Severity(), d.Summary(), d.Detail()))
	}

	if fwResp.Error != nil {
		//nolint:nilerr
		return toproto6.CallFunctionResponse(ctx, fwResp), nil
	}

	s.FrameworkServer.CallFunction(ctx, fwReq, fwResp)

	return toproto6.CallFunctionResponse(ctx, fwResp), nil
}
