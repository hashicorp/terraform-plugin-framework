package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func serveParseProviderMeta(ctx context.Context, provider Provider, dv *tfprotov6.DynamicValue) (ReadOnlyData, bool, diag.Diagnostics) {
	schema, diags := serveGetProviderMetaSchema(ctx, provider)
	if diags.HasError() {
		return ReadOnlyData{}, false, diags
	}
	if schema == nil {
		return ReadOnlyData{}, false, diags
	}
	res, ds := parseConfig(ctx, dv, schema)
	diags.Append(ds...)
	return res, true, diags
}
