package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeTwo struct{}

func (rt testServeResourceTypeTwo) GetSchema(_ context.Context) (Schema, []*tfprotov6.Diagnostic) {
	return Schema{
		Attributes: map[string]Attribute{
			"id": {
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"disks": {
				Optional: true,
				Computed: true,
				Attributes: ListNestedAttributes(map[string]Attribute{
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
				}, ListNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (rt testServeResourceTypeTwo) NewResource(_ context.Context, p Provider) (Resource, []*tfprotov6.Diagnostic) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceTwo{
		provider: provider,
	}, nil
}

var testServeResourceTypeTwoSchema = &tfprotov6.Schema{
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

var testServeResourceTypeTwoType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id": tftypes.String,
		"disks": tftypes.List{ElementType: tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"boot":    tftypes.Bool,
				"name":    tftypes.String,
				"size_gb": tftypes.Number,
			}},
		},
	},
}

type testServeResourceTwo struct {
	provider *testServeProvider
}

func (r testServeResourceTwo) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	r.provider.applyResourceChangePlannedStateValue = req.Plan.Raw
	r.provider.applyResourceChangePlannedStateSchema = req.Plan.Schema
	r.provider.applyResourceChangeConfigValue = req.Config.Raw
	r.provider.applyResourceChangeConfigSchema = req.Config.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_two"
	r.provider.applyResourceChangeCalledAction = "create"
	r.provider.createFunc(ctx, req, resp)
}

func (r testServeResourceTwo) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	r.provider.readResourceCurrentStateValue = req.State.Raw
	r.provider.readResourceCurrentStateSchema = req.State.Schema
	r.provider.readResourceProviderMetaValue = req.ProviderMeta.Raw
	r.provider.readResourceProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.readResourceCalledResourceType = "test_two"
	r.provider.readResourceImpl(ctx, req, resp)
}

func (r testServeResourceTwo) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	r.provider.applyResourceChangePriorStateValue = req.State.Raw
	r.provider.applyResourceChangePriorStateSchema = req.State.Schema
	r.provider.applyResourceChangePlannedStateValue = req.Plan.Raw
	r.provider.applyResourceChangePlannedStateSchema = req.Plan.Schema
	r.provider.applyResourceChangeConfigValue = req.Config.Raw
	r.provider.applyResourceChangeConfigSchema = req.Config.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_two"
	r.provider.applyResourceChangeCalledAction = "update"
	r.provider.updateFunc(ctx, req, resp)
}

func (r testServeResourceTwo) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	r.provider.applyResourceChangePriorStateValue = req.State.Raw
	r.provider.applyResourceChangePriorStateSchema = req.State.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_two"
	r.provider.applyResourceChangeCalledAction = "delete"
	r.provider.deleteFunc(ctx, req, resp)
}
