package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func ActionState(ctx context.Context, fw map[string]*tfsdk.State) (map[string]*tfprotov6.DynamicValue, diag.Diagnostics) {
	if fw == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	resp := make(map[string]*tfprotov6.DynamicValue)

	for s, state := range fw {
		data := &fwschemadata.Data{
			Description:    fwschemadata.DataDescriptionState,
			Schema:         state.Schema,
			TerraformValue: state.Raw,
		}
		valDiag := diag.Diagnostics{}
		resp[s], valDiag = DynamicValue(ctx, data)
		diags.Append(valDiag...)
	}

	return resp, diags
}
