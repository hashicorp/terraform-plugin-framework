package fromproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ValidateDataSourceConfigRequest returns the *fwserver.ValidateDataSourceConfigRequest
// equivalent of a *tfprotov5.ValidateDataSourceConfigRequest.
func ValidateDataSourceConfigRequest(ctx context.Context, proto5 *tfprotov5.ValidateDataSourceConfigRequest, dataSourceType tfsdk.DataSourceType, dataSourceSchema *tfsdk.Schema) (*fwserver.ValidateDataSourceConfigRequest, diag.Diagnostics) {
	if proto5 == nil {
		return nil, nil
	}

	fw := &fwserver.ValidateDataSourceConfigRequest{}

	config, diags := Config(ctx, proto5.Config, dataSourceSchema)

	fw.Config = config
	fw.DataSourceType = dataSourceType

	return fw, diags
}
