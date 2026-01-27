// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// DeleteStateRequest returns the *fwserver.DeleteStateRequest
// equivalent of a *tfprotov6.DeleteStateRequest.
func DeleteStateRequest(ctx context.Context, proto6 *tfprotov6.DeleteStateRequest) *fwserver.DeleteStatesRequest {
	if proto6 == nil {
		return nil
	}

	fw := &fwserver.DeleteStatesRequest{}

	return fw
}
