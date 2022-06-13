package fromproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ReadResourceRequest returns the *fwserver.ReadResourceRequest
// equivalent of a *tfprotov5.ReadResourceRequest.
func ReadResourceRequest(ctx context.Context, proto5 *tfprotov5.ReadResourceRequest, resourceType tfsdk.ResourceType, resourceSchema *tfsdk.Schema, providerMetaSchema *tfsdk.Schema) (*fwserver.ReadResourceRequest, diag.Diagnostics) {
	if proto5 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	fw := &fwserver.ReadResourceRequest{
		Private:      proto5.Private,
		ResourceType: resourceType,
	}

	currentState, currentStateDiags := State(ctx, proto5.CurrentState, resourceSchema)

	diags.Append(currentStateDiags...)

	fw.CurrentState = currentState

	providerMeta, providerMetaDiags := ProviderMeta(ctx, proto5.ProviderMeta, providerMetaSchema)

	diags.Append(providerMetaDiags...)

	fw.ProviderMeta = providerMeta

	return fw, diags
}
