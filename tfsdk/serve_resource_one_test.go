package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeOne struct{}

func (rt testServeResourceTypeOne) GetSchema(_ context.Context) (Schema, []*tfprotov6.Diagnostic) {
	return Schema{
		Version: 1,
		Attributes: map[string]Attribute{
			"name": {
				Required: true,
				Type:     types.StringType,
			},
			"favorite_colors": {
				Optional:      true,
				Type:          types.ListType{ElemType: types.StringType},
				PlanModifiers: RequiresReplace(),
			},
			"created_timestamp": {
				Computed: true,
				Type:     types.StringType,
			},
		},
	}, nil
}

func (rt testServeResourceTypeOne) NewResource(_ context.Context, p Provider) (Resource, []*tfprotov6.Diagnostic) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceOne{
		provider: provider,
	}, nil
}

var testServeResourceTypeOneSchema = &tfprotov6.Schema{
	Version: 1,
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "created_timestamp",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "favorite_colors",
				Optional: true,
				Type:     tftypes.List{ElementType: tftypes.String},
			},
			{
				Name:     "name",
				Required: true,
				Type:     tftypes.String,
			},
		},
	},
}

var testServeResourceTypeOneType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"created_timestamp": tftypes.String,
		"favorite_colors":   tftypes.List{ElementType: tftypes.String},
		"name":              tftypes.String,
	},
}

type testServeResourceOne struct {
	provider *testServeProvider
}

func (r testServeResourceOne) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	r.provider.applyResourceChangePlannedStateValue = req.Plan.Raw
	r.provider.applyResourceChangePlannedStateSchema = req.Plan.Schema
	r.provider.applyResourceChangeConfigValue = req.Config.Raw
	r.provider.applyResourceChangeConfigSchema = req.Config.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_one"
	r.provider.applyResourceChangeCalledAction = "create"
	r.provider.createFunc(ctx, req, resp)
}

func (r testServeResourceOne) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	r.provider.readResourceCurrentStateValue = req.State.Raw
	r.provider.readResourceCurrentStateSchema = req.State.Schema
	r.provider.readResourceProviderMetaValue = req.ProviderMeta.Raw
	r.provider.readResourceProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.readResourceCalledResourceType = "test_one"
	r.provider.readResourceImpl(ctx, req, resp)
}

func (r testServeResourceOne) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	r.provider.applyResourceChangePriorStateValue = req.State.Raw
	r.provider.applyResourceChangePriorStateSchema = req.State.Schema
	r.provider.applyResourceChangePlannedStateValue = req.Plan.Raw
	r.provider.applyResourceChangePlannedStateSchema = req.Plan.Schema
	r.provider.applyResourceChangeConfigValue = req.Config.Raw
	r.provider.applyResourceChangeConfigSchema = req.Config.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_one"
	r.provider.applyResourceChangeCalledAction = "update"
	r.provider.updateFunc(ctx, req, resp)
}

func (r testServeResourceOne) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	r.provider.applyResourceChangePriorStateValue = req.State.Raw
	r.provider.applyResourceChangePriorStateSchema = req.State.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_one"
	r.provider.applyResourceChangeCalledAction = "delete"
	r.provider.deleteFunc(ctx, req, resp)
}
