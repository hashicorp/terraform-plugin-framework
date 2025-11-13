// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

// ConfigureStateStoreRequest returns the *fwserver.ConfigureStateStoreRequest
// equivalent of a *tfprotov6.ConfigureStateStoreRequest.
func ConfigureStateStoreRequest(ctx context.Context, proto6 *tfprotov6.ConfigureStateStoreRequest, statestoreSchema fwschema.Schema) (*statestore.ConfigureStateStoreRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	fw := &statestore.ConfigureStateStoreRequest{
		TypeName:     proto6.TypeName,
		Capabilities: ConfigureStateStoreClientCapabilities(proto6.ClientCapabilities), //TODO: Add to plugin-go tfprotov6 and implement properly when capabilities are defined.
	}

	config, diags := Config(ctx, proto6.Config, statestoreSchema)

	if config != nil {
		fw.Config = *config
	}

	return fw, diags
}
