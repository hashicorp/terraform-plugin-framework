package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ValidateDataSourceConfigRequest is the framework server request for the
// ValidateDataSourceConfig RPC.
type ValidateDataSourceConfigRequest struct {
	Config         *tfsdk.Config
	DataSourceType tfsdk.DataSourceType
}

// ValidateDataSourceConfigResponse is the framework server response for the
// ValidateDataSourceConfig RPC.
type ValidateDataSourceConfigResponse struct {
	Diagnostics diag.Diagnostics
}

// ValidateDataSourceConfig implements the framework server ValidateDataSourceConfig RPC.
func (s *Server) ValidateDataSourceConfig(ctx context.Context, req *ValidateDataSourceConfigRequest, resp *ValidateDataSourceConfigResponse) {
	if req == nil || req.Config == nil {
		return
	}

	// Always instantiate new DataSource instances.
	logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType NewDataSource")
	dataSource, diags := req.DataSourceType.NewDataSource(ctx, s.Provider)
	logging.FrameworkDebug(ctx, "Called provider defined DataSourceType NewDataSource")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	vdscReq := tfsdk.ValidateDataSourceConfigRequest{
		Config: *req.Config,
	}

	if dataSource, ok := dataSource.(tfsdk.DataSourceWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "DataSource implements DataSourceWithConfigValidators")

		for _, configValidator := range dataSource.ConfigValidators(ctx) {
			vdscResp := &tfsdk.ValidateDataSourceConfigResponse{
				Diagnostics: resp.Diagnostics,
			}

			logging.FrameworkDebug(
				ctx,
				"Calling provider defined DataSourceConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)
			configValidator.Validate(ctx, vdscReq, vdscResp)
			logging.FrameworkDebug(
				ctx,
				"Called provider defined DataSourceConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)

			resp.Diagnostics = vdscResp.Diagnostics
		}
	}

	if dataSource, ok := dataSource.(tfsdk.DataSourceWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "DataSource implements DataSourceWithValidateConfig")

		vdscResp := &tfsdk.ValidateDataSourceConfigResponse{
			Diagnostics: resp.Diagnostics,
		}

		logging.FrameworkDebug(ctx, "Calling provider defined DataSource ValidateConfig")
		dataSource.ValidateConfig(ctx, vdscReq, vdscResp)
		logging.FrameworkDebug(ctx, "Called provider defined DataSource ValidateConfig")

		resp.Diagnostics = vdscResp.Diagnostics
	}

	validateSchemaReq := ValidateSchemaRequest{
		Config: *req.Config,
	}
	validateSchemaResp := ValidateSchemaResponse{
		Diagnostics: resp.Diagnostics,
	}

	SchemaValidate(ctx, req.Config.Schema, validateSchemaReq, &validateSchemaResp)

	resp.Diagnostics = validateSchemaResp.Diagnostics
}
