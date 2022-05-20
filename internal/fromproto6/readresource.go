package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ReadResourceRequest returns the *fwserver.ReadResourceRequest
// equivalent of a *tfprotov6.ReadResourceRequest.
func ReadResourceRequest(ctx context.Context, proto6 *tfprotov6.ReadResourceRequest, resourceType tfsdk.ResourceType, resourceSchema *tfsdk.Schema, providerMetaSchema *tfsdk.Schema) (*fwserver.ReadResourceRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	fw := &fwserver.ReadResourceRequest{
		Private:      proto6.Private,
		ResourceType: resourceType,
	}

	currentState, currentStateDiags := State(ctx, proto6.CurrentState, resourceSchema)

	diags.Append(currentStateDiags...)

	fw.CurrentState = currentState

	providerMeta, providerMetaDiags := ProviderMeta(ctx, proto6.ProviderMeta, providerMetaSchema)

	diags.Append(providerMetaDiags...)

	fw.ProviderMeta = providerMeta

	return fw, diags
}
