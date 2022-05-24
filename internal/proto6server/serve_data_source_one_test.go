package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeDataSourceTypeOne struct{}

func (dt testServeDataSourceTypeOne) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"current_time": {
				Type:     types.StringType,
				Computed: true,
			},
			"current_date": {
				Type:     types.StringType,
				Computed: true,
			},
			"is_dst": {
				Type:     types.BoolType,
				Computed: true,
			},
		},
	}, nil
}

func (dt testServeDataSourceTypeOne) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeDataSourceOne{
		provider: provider,
	}, nil
}

var testServeDataSourceTypeOneType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"current_date": tftypes.String,
		"current_time": tftypes.String,
		"is_dst":       tftypes.Bool,
	},
}

type testServeDataSourceOne struct {
	provider *testServeProvider
}

func (r testServeDataSourceOne) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	r.provider.readDataSourceConfigValue = req.Config.Raw
	r.provider.readDataSourceConfigSchema = req.Config.Schema
	r.provider.readDataSourceProviderMetaValue = req.ProviderMeta.Raw
	r.provider.readDataSourceProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.readDataSourceCalledDataSourceType = "test_one"
	r.provider.readDataSourceImpl(ctx, req, resp)
}
