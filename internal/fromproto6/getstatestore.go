// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// GetStatesRequest returns the *fwserver.GetStatesRequest
// equivalent of a *tfprotov6.GetStatesRequest.
func GetStatesRequest(ctx context.Context, proto6 *tfprotov6.GetStatesRequest) *fwserver.GetStatesRequest {
	if proto6 == nil {
		return nil
	}

	fw := &fwserver.GetStatesRequest{}

	return fw
}
