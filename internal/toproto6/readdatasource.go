// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

// ReadDataSourceResponse returns the *tfprotov6.ReadDataSourceResponse
// equivalent of a *fwserver.ReadDataSourceResponse.
func ReadDataSourceResponse(ctx context.Context, fw *fwserver.ReadDataSourceResponse) *tfprotov6.ReadDataSourceResponse {
	if fw == nil {
		return nil
	}

	proto6 := &tfprotov6.ReadDataSourceResponse{
		Diagnostics: Diagnostics(ctx, fw.Diagnostics),
	}

	state, diags := State(ctx, fw.State)

	proto6.Diagnostics = append(proto6.Diagnostics, Diagnostics(ctx, diags)...)
	proto6.State = state

	if fw.Deferred != nil {
		proto6.Deferred = &tfprotov6.Deferred{
			Reason: tfprotov6.DeferredReason(fw.Deferred.Reason),
		}
	}

	return proto6
}
