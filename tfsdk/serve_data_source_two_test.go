package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeDataSourceTypeTwo struct{}

func (dt testServeDataSourceTypeTwo) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"family": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (dt testServeDataSourceTypeTwo) NewDataSource(_ context.Context, p Provider) (DataSource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeDataSourceTwo{
		provider: provider,
	}, nil
}

var testServeDataSourceTypeTwoSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "family",
				Optional: true,
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "id",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "name",
				Optional: true,
				Computed: true,
				Type:     tftypes.String,
			},
		},
	},
}

var testServeDataSourceTypeTwoType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"family": tftypes.String,
		"name":   tftypes.String,
		"id":     tftypes.String,
	},
}

type testServeDataSourceTwo struct {
	provider *testServeProvider
}

func (r testServeDataSourceTwo) Read(ctx context.Context, req ReadDataSourceRequest, resp *ReadDataSourceResponse) {
	r.provider.readDataSourceConfigValue = req.Config.Raw
	r.provider.readDataSourceConfigSchema = req.Config.Schema
	r.provider.readDataSourceProviderMetaValue = req.ProviderMeta.Raw
	r.provider.readDataSourceProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.readDataSourceCalledDataSourceType = "test_two"
	r.provider.readDataSourceImpl(ctx, req, resp)
}
