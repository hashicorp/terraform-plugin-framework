package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeDataSourceTwo struct{}

func (dt testServeDataSourceTwo) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
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

func (dt testServeDataSourceTwo) NewDataSource(_ Provider) (DataSource, []*tfprotov6.Diagnostic) {
	panic("not implemented") // TODO: Implement
}

var testServeDataSourceTwoSchema = &tfprotov6.Schema{
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
