package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ReadDataSourceRequest returns the *fwserver.ReadDataSourceRequest
// equivalent of a *tfprotov6.ReadDataSourceRequest.
func ReadDataSourceRequest(ctx context.Context, proto6 *tfprotov6.ReadDataSourceRequest, dataSourceType tfsdk.DataSourceType, dataSourceSchema *tfsdk.Schema, providerMetaSchema *tfsdk.Schema) (*fwserver.ReadDataSourceRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	fw := &fwserver.ReadDataSourceRequest{
		DataSourceType: dataSourceType,
	}

	config, configDiags := Config(ctx, proto6.Config, dataSourceSchema)

	diags.Append(configDiags...)

	fw.Config = config

	providerMeta, providerMetaDiags := ProviderMeta(ctx, proto6.ProviderMeta, providerMetaSchema)

	diags.Append(providerMetaDiags...)

	fw.ProviderMeta = providerMeta

	return fw, diags
}
