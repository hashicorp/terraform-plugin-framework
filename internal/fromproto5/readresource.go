package fromproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// ReadResourceRequest returns the *fwserver.ReadResourceRequest
// equivalent of a *tfprotov5.ReadResourceRequest.
func ReadResourceRequest(ctx context.Context, proto5 *tfprotov5.ReadResourceRequest, resourceType provider.ResourceType, resourceSchema fwschema.Schema, providerMetaSchema fwschema.Schema) (*fwserver.ReadResourceRequest, diag.Diagnostics) {
	if proto5 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	fw := &fwserver.ReadResourceRequest{
		ResourceType: resourceType,
	}

	currentState, currentStateDiags := State(ctx, proto5.CurrentState, resourceSchema)

	diags.Append(currentStateDiags...)

	fw.CurrentState = currentState

	providerMeta, providerMetaDiags := ProviderMeta(ctx, proto5.ProviderMeta, providerMetaSchema)

	diags.Append(providerMetaDiags...)

	fw.ProviderMeta = providerMeta

	privateData, privateDataDiags := PrivateData(ctx, proto5.Private)

	diags.Append(privateDataDiags...)

	fw.Private = privateData

	return fw, diags
}
