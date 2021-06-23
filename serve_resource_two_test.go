package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTwo struct{}

func (rt testServeResourceTwo) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": {
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"disks": {
				Optional: true,
				Computed: true,
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"name": {
						Required: true,
						Type:     types.StringType,
					},
					"size_gb": {
						Required: true,
						Type:     types.NumberType,
					},
					"boot": {
						Required: true,
						Type:     types.BoolType,
					},
				}, schema.ListNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (rt testServeResourceTwo) NewResource(_ Provider) (Resource, []*tfprotov6.Diagnostic) {
	panic("not implemented") // TODO: Implement
}

var testServeResourceTwoSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "disks",
				Optional: true,
				Computed: true,
				NestedType: &tfprotov6.SchemaObject{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "boot",
							Required: true,
							Type:     tftypes.Bool,
						},
						{
							Name:     "name",
							Required: true,
							Type:     tftypes.String,
						},
						{
							Name:     "size_gb",
							Required: true,
							Type:     tftypes.Number,
						},
					},
					Nesting: tfprotov6.SchemaObjectNestingModeList,
				},
			},
			{
				Name:     "id",
				Optional: true,
				Computed: true,
				Type:     tftypes.String,
			},
		},
	},
}
