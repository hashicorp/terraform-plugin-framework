package fromproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ReadDataSourceRequest returns the *fwserver.ReadDataSourceRequest
// equivalent of a *tfprotov5.ReadDataSourceRequest.
func ReadDataSourceRequest(ctx context.Context, proto5 *tfprotov5.ReadDataSourceRequest, dataSourceType tfsdk.DataSourceType, dataSourceSchema *tfsdk.Schema, providerMetaSchema *tfsdk.Schema) (*fwserver.ReadDataSourceRequest, diag.Diagnostics) {
	if proto5 == nil {
		return nil, nil
	}

	var diags diag.Diagnostics

	// Panic prevention here to simplify the calling implementations.
	// This should not happen, but just in case.
	if dataSourceSchema == nil {
		diags.AddError(
			"Missing DataSource Schema",
			"An unexpected error was encountered when handling the request. "+
				"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
				"Please report this to the provider developer:\n\n"+
				"Missing schema.",
		)

		return nil, diags
	}

	fw := &fwserver.ReadDataSourceRequest{
		DataSourceSchema: *dataSourceSchema,
		DataSourceType:   dataSourceType,
	}

	config, configDiags := Config(ctx, proto5.Config, dataSourceSchema)

	diags.Append(configDiags...)

	fw.Config = config

	providerMeta, providerMetaDiags := ProviderMeta(ctx, proto5.ProviderMeta, providerMetaSchema)

	diags.Append(providerMetaDiags...)

	fw.ProviderMeta = providerMeta

	return fw, diags
}
