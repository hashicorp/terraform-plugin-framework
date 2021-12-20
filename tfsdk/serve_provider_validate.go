package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// serveValidateProviderWithConfigValidators runs any declarative validators
// that the provider may declare.
func serveValidateProviderWithConfigValidators(ctx context.Context, provider Provider, config ReadOnlyData, diags diag.Diagnostics) diag.Diagnostics {
	if vpc, ok := provider.(ProviderWithConfigValidators); ok {
		for _, configValidator := range vpc.ConfigValidators(ctx) {
			vpcReq := ValidateProviderConfigRequest{
				Config: config,
			}
			vpcRes := &ValidateProviderConfigResponse{
				Diagnostics: diags,
			}

			configValidator.Validate(ctx, vpcReq, vpcRes)

			diags = vpcRes.Diagnostics
		}
	}
	return diags
}

// serveValidateProviderWithValidateConfig runs the imperative validator at the
// provider level, if one is defined.
func serveValidateProviderWithValidateConfig(ctx context.Context, provider Provider, config ReadOnlyData, diags diag.Diagnostics) diag.Diagnostics {
	if vpc, ok := provider.(ProviderWithValidateConfig); ok {
		vpcReq := ValidateProviderConfigRequest{
			Config: config,
		}
		vpcRes := &ValidateProviderConfigResponse{
			Diagnostics: diags,
		}

		vpc.ValidateConfig(ctx, vpcReq, vpcRes)

		diags = vpcRes.Diagnostics
	}
	return diags
}
