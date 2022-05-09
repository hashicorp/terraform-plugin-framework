package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// ValidateProviderConfigRequest is the framework server request for the
// ValidateProviderConfig RPC.
type ValidateProviderConfigRequest struct {
	Config *tfsdk.Config
}

// ValidateProviderConfigResponse is the framework server response for the
// ValidateProviderConfig RPC.
type ValidateProviderConfigResponse struct {
	PreparedConfig *tfsdk.Config
	Diagnostics    diag.Diagnostics
}

// ValidateProviderConfig implements the framework server ValidateProviderConfig RPC.
func (s *Server) ValidateProviderConfig(ctx context.Context, req *ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {
	if req == nil || req.Config == nil {
		return
	}

	vpcReq := tfsdk.ValidateProviderConfigRequest{
		Config: *req.Config,
	}

	if provider, ok := s.Provider.(tfsdk.ProviderWithConfigValidators); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithConfigValidators")

		for _, configValidator := range provider.ConfigValidators(ctx) {
			vpcRes := &tfsdk.ValidateProviderConfigResponse{
				Diagnostics: resp.Diagnostics,
			}

			logging.FrameworkDebug(
				ctx,
				"Calling provider defined ProviderConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)
			configValidator.Validate(ctx, vpcReq, vpcRes)
			logging.FrameworkDebug(
				ctx,
				"Called provider defined ProviderConfigValidator",
				map[string]interface{}{
					logging.KeyDescription: configValidator.Description(ctx),
				},
			)

			resp.Diagnostics = vpcRes.Diagnostics
		}
	}

	if provider, ok := s.Provider.(tfsdk.ProviderWithValidateConfig); ok {
		logging.FrameworkTrace(ctx, "Provider implements ProviderWithValidateConfig")

		vpcRes := &tfsdk.ValidateProviderConfigResponse{
			Diagnostics: resp.Diagnostics,
		}

		logging.FrameworkDebug(ctx, "Calling provider defined Provider ValidateConfig")
		provider.ValidateConfig(ctx, vpcReq, vpcRes)
		logging.FrameworkDebug(ctx, "Called provider defined Provider ValidateConfig")

		resp.Diagnostics = vpcRes.Diagnostics
	}

	validateSchemaReq := ValidateSchemaRequest{
		Config: *req.Config,
	}
	validateSchemaResp := ValidateSchemaResponse{
		Diagnostics: resp.Diagnostics,
	}

	SchemaValidate(ctx, req.Config.Schema, validateSchemaReq, &validateSchemaResp)

	resp.Diagnostics = validateSchemaResp.Diagnostics

	// This RPC allows a modified configuration to be returned. This was
	// previously used to allow a "required" provider attribute (as defined
	// by a schema) to still be "optional" with a default value, typically
	// through an environment variable. Other tooling based on the provider
	// schema information could not determine this implementation detail.
	// To ensure accuracy going forward, this implementation is opinionated
	// towards accurate provider schema definitions and optional values
	// can be filled in or return errors during ConfigureProvider().
	resp.PreparedConfig = req.Config
}
