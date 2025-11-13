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

// ValidateStateStoreRequest returns the *fwserver.ValidateStateStoreRequest
// equivalent of a *tfprotov6.ValidateStateStoreRequest.
func ValidateStateStoreRequest(ctx context.Context, proto6 *tfprotov6.ValidateStateStoreRequest, reqStateStore statestore.StateStore, StateStoreSchema fwschema.Schema) (*fwserver.ValidateStateStoreRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	fw := &fwserver.ValidateStateStoreRequest{}

	config, diags := Config(ctx, proto6.Config, StateStoreSchema)

	fw.Config = config
	fw.StateStore = reqStateStore

	return fw, diags
}
