// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// PlanActionRequest returns the *fwserver.PlanActionRequest equivalent of a *tfprotov5.PlanActionRequest.
func PlanActionRequest(ctx context.Context, proto5 *tfprotov5.PlanActionRequest, reqAction action.Action, actionSchema fwschema.Schema, linkedResourceSchemas []fwschema.Schema, linkedResourceIdentitySchemas []fwschema.Schema) (*fwserver.PlanActionRequest, diag.Diagnostics) {
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

	if len(proto5.LinkedResources) != len(linkedResourceSchemas) {
		diags.AddError(
			"Mismatched Linked Resources in PlanAction Request",
			"An unexpected error was encountered when handling the request. "+
				"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.\n\n"+
				fmt.Sprintf(
					"Received %d linked resource(s), but the provider was expecting %d linked resource(s).",
					len(proto5.LinkedResources),
					len(linkedResourceSchemas),
				),
		)

		return nil, diags
	}

	// MAINTAINER NOTE: The number of identity schemas should always be in sync (if not supported, will have nil),
	// so this error check is more for panic prevention.
	if len(proto5.LinkedResources) != len(linkedResourceIdentitySchemas) {
		diags.AddError(
			"Mismatched Linked Resources in PlanAction Request",
			"An unexpected error was encountered when handling the request. "+
				"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.\n\n"+
				fmt.Sprintf(
					"Received %d linked resource(s), but the provider was expecting %d linked resource(s).",
					len(proto5.LinkedResources),
					len(linkedResourceIdentitySchemas),
				),
		)

		return nil, diags
	}

	for i, linkedResource := range proto5.LinkedResources {
		schema := linkedResourceSchemas[i]
		identitySchema := linkedResourceIdentitySchemas[i]

		// Config
		config, configDiags := Config(ctx, linkedResource.Config, schema)
		diags.Append(configDiags...)

		// Prior state
		priorState, priorStateDiags := State(ctx, linkedResource.PriorState, schema)
		diags.Append(priorStateDiags...)

		// Planned state (plan)
		plannedState, plannedStateDiags := Plan(ctx, linkedResource.PlannedState, schema)
		diags.Append(plannedStateDiags...)

		// Prior identity
		var priorIdentity *tfsdk.ResourceIdentity
		if linkedResource.PriorIdentity != nil {
			if identitySchema == nil {
				// MAINTAINER NOTE: Not all linked resources support identity, so it's valid for an identity schema to be nil. However,
				// it's not valid for Terraform core to send an identity for a linked resource that doesn't support one. This would likely indicate
				// that there is a bug in the definition of the linked resources (not including an identity schema when it is supported), or a bug in
				// either Terraform core/Framework.
				diags.AddError(
					"Unable to Convert Linked Resource Identity",
					"An unexpected error was encountered when converting a linked resource identity from the protocol type. "+
						fmt.Sprintf("Linked resource (at index %d) contained identity data, but the resource doesn't support identity.\n\n", i)+
						"This is always a problem with the provider and should be reported to the provider developer.",
				)
				return nil, diags
			}

			identityVal, priorIdentityDiags := ResourceIdentity(ctx, linkedResource.PriorIdentity, identitySchema)
			diags.Append(priorIdentityDiags...)

			priorIdentity = identityVal
		}

		fw.LinkedResources = append(fw.LinkedResources, &fwserver.PlanActionLinkedResourceRequest{
			Config:        config,
			PlannedState:  plannedState,
			PriorState:    priorState,
			PriorIdentity: priorIdentity,
		})
	}

	return fw, diags
}
