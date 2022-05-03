package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeProviderWithConfigValidators struct {
	*testServeProvider
}

func (t *testServeProviderWithConfigValidators) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

var testServeProviderWithConfigValidatorsType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"string": tftypes.String,
	},
}

func (p testServeProviderWithConfigValidators) ConfigValidators(ctx context.Context) []tfsdk.ProviderConfigValidator {
	return []tfsdk.ProviderConfigValidator{
		newTestProviderConfigValidator(p.validateProviderConfigImpl),
		// Verify multiple validators
		newTestProviderConfigValidator(p.validateProviderConfigImpl),
	}
}

type testProviderConfigValidator struct {
	tfsdk.ProviderConfigValidator

	impl func(context.Context, tfsdk.ValidateProviderConfigRequest, *tfsdk.ValidateProviderConfigResponse)
}

func (v testProviderConfigValidator) Description(ctx context.Context) string {
	return "test provider config validator"
}
func (v testProviderConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "**test** provider config validator"
}
func (v testProviderConfigValidator) Validate(ctx context.Context, req tfsdk.ValidateProviderConfigRequest, resp *tfsdk.ValidateProviderConfigResponse) {
	v.impl(ctx, req, resp)
}

func newTestProviderConfigValidator(impl func(context.Context, tfsdk.ValidateProviderConfigRequest, *tfsdk.ValidateProviderConfigResponse)) testProviderConfigValidator {
	return testProviderConfigValidator{impl: impl}
}
