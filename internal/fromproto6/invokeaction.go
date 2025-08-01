// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

// InvokeActionRequest returns the *fwserver.InvokeActionRequest equivalent of a *tfprotov6.InvokeActionRequest.
func InvokeActionRequest(ctx context.Context, proto6 *tfprotov6.InvokeActionRequest, reqAction action.Action, actionSchema fwschema.Schema) (*fwserver.InvokeActionRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	// Panic prevention here to simplify the calling implementations.
	// This should not happen, but just in case.
	if actionSchema == nil {
		diags.AddError(
			"Missing Action Schema",
			"An unexpected error was encountered when handling the request. "+
				"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
				"Please report this to the provider developer:\n\n"+
				"Missing schema.",
		)

		return nil, diags
	}

	fw := &fwserver.InvokeActionRequest{
		Action:       reqAction,
		ActionSchema: actionSchema,
	}

	config, configDiags := Config(ctx, proto6.Config, actionSchema)

	diags.Append(configDiags...)

	fw.Config = config

	// TODO:Actions: Here we need to retrieve linked resource data

	return fw, diags
}
