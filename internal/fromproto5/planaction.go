// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

// PlanActionRequest returns the *fwserver.PlanActionRequest equivalent of a *tfprotov5.PlanActionRequest.
func PlanActionRequest(ctx context.Context, proto5 *tfprotov5.PlanActionRequest, reqAction action.Action, actionSchema fwschema.Schema) (*fwserver.PlanActionRequest, diag.Diagnostics) {
	if proto5 == nil {
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

	fw := &fwserver.PlanActionRequest{
		Action:             reqAction,
		ActionSchema:       actionSchema,
		ClientCapabilities: ModifyPlanActionClientCapabilities(proto5.ClientCapabilities),
	}

	config, configDiags := Config(ctx, proto5.Config, actionSchema)

	diags.Append(configDiags...)

	fw.Config = config

	// TODO:Actions: Here we need to retrieve linked resource data

	return fw, diags
}
