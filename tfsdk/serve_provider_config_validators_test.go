package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeProviderWithConfigValidators struct {
	*testServeProvider
}

func (t *testServeProviderWithConfigValidators) GetSchema(_ context.Context) (Schema, []*tfprotov6.Diagnostic) {
	return Schema{
		Attributes: map[string]Attribute{
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

func (p testServeProviderWithConfigValidators) ConfigValidators(ctx context.Context) []ProviderConfigValidator {
	return []ProviderConfigValidator{
		newTestProviderConfigValidator(p.validateProviderConfigImpl),
	}
}

type testProviderConfigValidator struct {
	ProviderConfigValidator

	impl func(context.Context, ValidateProviderConfigRequest, *ValidateProviderConfigResponse)
}

func (v testProviderConfigValidator) Description(ctx context.Context) string {
	return "test provider config validator"
}
func (v testProviderConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "**test** provider config validator"
}
func (v testProviderConfigValidator) Validate(ctx context.Context, req ValidateProviderConfigRequest, resp *ValidateProviderConfigResponse) {
	v.impl(ctx, req, resp)
}

func newTestProviderConfigValidator(impl func(context.Context, ValidateProviderConfigRequest, *ValidateProviderConfigResponse)) testProviderConfigValidator {
	return testProviderConfigValidator{impl: impl}
}
