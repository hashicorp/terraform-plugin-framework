package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeDataSourceOne struct{}

func (dt testServeDataSourceOne) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
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

func (dt testServeDataSourceOne) NewDataSource(_ Provider) (DataSource, []*tfprotov6.Diagnostic) {
	panic("not implemented") // TODO: Implement
}

var testServeDataSourceOneSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "current_date",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "current_time",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "is_dst",
				Computed: true,
				Type:     tftypes.Bool,
			},
		},
	},
}
