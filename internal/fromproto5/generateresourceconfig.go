package fromproto5

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

func GenerateResourceConfigRequest(ctx context.Context, proto5 *tfprotov5.GenerateResourceConfigRequest, resourceSchema fwschema.Schema) (*fwserver.GenerateResourceConfigRequest, diag.Diagnostics) {
	if proto5 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	// TODO nil check and error
	state, stateDiags := State(ctx, proto5.State, resourceSchema)

	diags.Append(stateDiags...)

	fw := &fwserver.GenerateResourceConfigRequest{
		TypeName: proto5.TypeName,
		State:    state,
	}

	return fw, diags
}
