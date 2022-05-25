package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeConfigValidators struct{}

func (dt testServeResourceTypeConfigValidators) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (dt testServeResourceTypeConfigValidators) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceConfigValidators{
		provider: provider,
	}, nil
}

var testServeResourceTypeConfigValidatorsType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"string": tftypes.String,
	},
}

type testServeResourceConfigValidators struct {
	provider *testServeProvider
}

func (r testServeResourceConfigValidators) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceConfigValidators) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceConfigValidators) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceConfigValidators) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceConfigValidators) ConfigValidators(ctx context.Context) []tfsdk.ResourceConfigValidator {
	r.provider.validateResourceConfigCalledResourceType = "test_config_validators"

	return []tfsdk.ResourceConfigValidator{
		newTestResourceConfigValidator(r.provider.validateResourceConfigImpl),
		// Verify multiple validators
		newTestResourceConfigValidator(r.provider.validateResourceConfigImpl),
	}
}

type testResourceConfigValidator struct {
	tfsdk.ResourceConfigValidator

	impl func(context.Context, tfsdk.ValidateResourceConfigRequest, *tfsdk.ValidateResourceConfigResponse)
}

func (v testResourceConfigValidator) Description(ctx context.Context) string {
	return "test resource config validator"
}
func (v testResourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "**test** resource config validator"
}
func (v testResourceConfigValidator) Validate(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	v.impl(ctx, req, resp)
}

func newTestResourceConfigValidator(impl func(context.Context, tfsdk.ValidateResourceConfigRequest, *tfsdk.ValidateResourceConfigResponse)) testResourceConfigValidator {
	return testResourceConfigValidator{impl: impl}
}
