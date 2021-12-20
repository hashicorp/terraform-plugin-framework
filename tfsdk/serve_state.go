package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// parseState turns a *tfprotov6.DynamicValue and a Schema into a State.
func parseState(ctx context.Context, dv *tfprotov6.DynamicValue, schema Schema) (*Data, diag.Diagnostics) {
	var diags diag.Diagnostics
	state, err := dv.Unmarshal(schema.TerraformType(ctx))
	if err != nil {
		diags.AddError(
			"Error parsing state",
			"The provider had a problem parsing the state. Report this to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}
	obj, err := objectFromSchemaAndTerraformValue(ctx, schema, state)
	if err != nil {
		// TODO: return error
	}
	return &Data{
		ReadOnlyData: ReadOnlyData{
			Values: obj,
		},
	}, diags
}
