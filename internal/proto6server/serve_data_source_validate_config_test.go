package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeDataSourceTypeValidateConfig struct{}

func (dt testServeDataSourceTypeValidateConfig) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (dt testServeDataSourceTypeValidateConfig) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeDataSourceValidateConfig{
		provider: provider,
	}, nil
}

var testServeDataSourceTypeValidateConfigType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"string": tftypes.String,
	},
}

type testServeDataSourceValidateConfig struct {
	provider *testServeProvider
}

func (r testServeDataSourceValidateConfig) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeDataSourceValidateConfig) ValidateConfig(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
	r.provider.validateDataSourceConfigCalledDataSourceType = "test_validate_config"
	r.provider.validateDataSourceConfigImpl(ctx, req, resp)
}
