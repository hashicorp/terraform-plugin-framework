// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ReadStateBytesRequest returns the *fwserver.ReadStateBytesRequest
// equivalent of a *tfprotov6.ReadStateBytesRequest.
func ReadStateBytesRequest(ctx context.Context, proto6 *tfprotov6.ReadStateBytesRequest, stateStore statestore.StateStore, statestoreSchema fwschema.Schema) (*fwserver.ReadStateBytesRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	if proto6.StateID == "" {
		diags.AddError(
			"Missing State ID",
			"An unexpected error was encountered when handling the request. "+
				"This is always an issue in terraform-plugin-framework or Terraform itself and should be reported to the provider developers.\n\n"+
				"Please report this to the provider developer:\n\n"+
				"Missing State ID.",
		)

		return nil, diags
	}

	fw := &fwserver.ReadStateBytesRequest{
		StateStore: stateStore,
		StateID:    proto6.StateID,
	}

	return fw, diags
}
