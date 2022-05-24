package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeValidateConfig struct{}

func (dt testServeResourceTypeValidateConfig) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (dt testServeResourceTypeValidateConfig) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceValidateConfig{
		provider: provider,
	}, nil
}

var testServeResourceTypeValidateConfigType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"string": tftypes.String,
	},
}

type testServeResourceValidateConfig struct {
	provider *testServeProvider
}

func (r testServeResourceValidateConfig) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceValidateConfig) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceValidateConfig) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceValidateConfig) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceValidateConfig) ValidateConfig(ctx context.Context, req tfsdk.ValidateResourceConfigRequest, resp *tfsdk.ValidateResourceConfigResponse) {
	r.provider.validateResourceConfigCalledResourceType = "test_validate_config"
	r.provider.validateResourceConfigImpl(ctx, req, resp)
}
