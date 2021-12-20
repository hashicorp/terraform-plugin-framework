package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// parseConfig turns a *tfprotov6.DynamicValue and a Schema into a Config.
func parseConfig(ctx context.Context, dv *tfprotov6.DynamicValue, schema Schema) (Config, diag.Diagnostics) {
	var diags diag.Diagnostics
	conf, err := dv.Unmarshal(schema.TerraformType(ctx))
	if err != nil {
		diags.AddError(
			"Error parsing config",
			"The provider had a problem parsing the config. Report this to the provider developer:\n\n"+err.Error(),
		)
		return Config{}, diags
	}
	return Config{
		Raw:    conf,
		Schema: schema,
	}, diags
}
