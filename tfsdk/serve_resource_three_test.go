package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func (rt testServeResourceTypeThree) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Version: 1,
		Attributes: map[string]Attribute{
			"name": {
				Required: true,
				Type:     types.StringType,
			},
			"last_updated": {
				Computed: true,
				Type:     types.StringType,
			},
			"first_updated": {
				Computed: true,
				Type:     types.StringType,
			},
			"map_nested": {
				Required: true,
				Attributes: MapNestedAttributes(map[string]Attribute{
					"computed_string": {
						Computed: true,
						Type:     types.StringType,
					},
					"string": {
						Optional: true,
						Type:     types.StringType,
					},
				}, MapNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (rt testServeResourceTypeThree) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceThree{
		provider: provider,
	}, nil
}

var testServeResourceTypeThreeSchema = &tfprotov6.Schema{
	Version: 1,
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "first_updated",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "last_updated",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "map_nested",
				Required: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeMap,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "computed_string",
							Computed: true,
							Type:     tftypes.String,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
			},
			{
				Name:     "name",
				Required: true,
				Type:     tftypes.String,
			},
		},
	},
}

var testServeResourceTypeThreeType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"name":          tftypes.String,
		"last_updated":  tftypes.String,
		"first_updated": tftypes.String,
		"map_nested": tftypes.Map{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"computed_string": tftypes.String,
					"string":          tftypes.String,
				},
			},
		},
	},
}

type testServeResourceThree struct {
	provider *testServeProvider
}

type testServeResourceTypeThree struct{}

func (r testServeResourceThree) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceThree) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceThree) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceThree) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceThree) ImportState(ctx context.Context, req ImportResourceStateRequest, resp *ImportResourceStateResponse) {
	ResourceImportStateNotImplemented(ctx, "Not expected to be called during testing.", resp)
}
