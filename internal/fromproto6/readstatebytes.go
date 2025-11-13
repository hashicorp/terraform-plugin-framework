// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ReadStateBytesRequest returns the *fwserver.ReadStateBytesRequest
// equivalent of a *tfprotov6.ReadStateBytesRequest.
func ReadStateBytesRequest(ctx context.Context, proto6 *tfprotov6.ReadStateBytesRequest, stateBytes statestore.StateStore) (*fwserver.ReadStateBytesRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	// Panic prevention here to simplify the calling implementations.
	// This should not happen, but just in case.

	fw := &fwserver.ReadStateBytesRequest{
		StateId: proto6.StateId,
	}

	fw.StateId = proto6.StateId

	return fw, diags
}
