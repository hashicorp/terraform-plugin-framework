package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ValidateDataSourceConfigRequest returns the *fwserver.ValidateDataSourceConfigRequest
// equivalent of a *tfprotov6.ValidateDataSourceConfigRequest.
func ValidateDataSourceConfigRequest(ctx context.Context, proto6 *tfprotov6.ValidateDataResourceConfigRequest, dataSourceType tfsdk.DataSourceType, dataSourceSchema *tfsdk.Schema) (*fwserver.ValidateDataSourceConfigRequest, diag.Diagnostics) {
	if proto6 == nil {
		return nil, nil
	}

	fw := &fwserver.ValidateDataSourceConfigRequest{}

	config, diags := Config(ctx, proto6.Config, dataSourceSchema)

	fw.Config = config
	fw.DataSourceType = dataSourceType

	return fw, diags
}
