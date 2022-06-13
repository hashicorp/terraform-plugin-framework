package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ApplyResourceChangeRequest returns the *fwserver.ApplyResourceChangeRequest
// equivalent of a *tfprotov6.ApplyResourceChangeRequest.
func ApplyResourceChangeRequest(ctx context.Context, proto6 *tfprotov6.ApplyResourceChangeRequest, resourceType tfsdk.ResourceType, resourceSchema *tfsdk.Schema, providerMetaSchema *tfsdk.Schema) (*fwserver.ApplyResourceChangeRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	// Panic prevention here to simplify the calling implementations.
	// This should not happen, but just in case.
	if resourceSchema == nil {
		diags.AddError(
			"Missing Resource Schema",
			"An unexpected error was encountered when handling the request. "+
				"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
				"Please report this to the provider developer:\n\n"+
				"Missing schema.",
		)

		return nil, diags
	}

	fw := &fwserver.ApplyResourceChangeRequest{
		PlannedPrivate: proto6.PlannedPrivate,
		ResourceSchema: *resourceSchema,
		ResourceType:   resourceType,
	}

	config, configDiags := Config(ctx, proto6.Config, resourceSchema)

	diags.Append(configDiags...)

	fw.Config = config

	plannedState, plannedStateDiags := Plan(ctx, proto6.PlannedState, resourceSchema)

	diags.Append(plannedStateDiags...)

	fw.PlannedState = plannedState

	priorState, priorStateDiags := State(ctx, proto6.PriorState, resourceSchema)

	diags.Append(priorStateDiags...)

	fw.PriorState = priorState

	providerMeta, providerMetaDiags := ProviderMeta(ctx, proto6.ProviderMeta, providerMetaSchema)

	diags.Append(providerMetaDiags...)

	fw.ProviderMeta = providerMeta

	return fw, diags
}
