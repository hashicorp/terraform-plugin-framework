package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeTwo struct{}

func (rt testServeResourceTypeTwo) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"disks": {
				Optional: true,
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
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
				}),
			},
		},
		Blocks: map[string]tfsdk.Block{
			"list_nested_blocks": {
				Attributes: map[string]tfsdk.Attribute{
					"required_bool": {
						Required: true,
						Type:     types.BoolType,
					},
					"required_number": {
						Required: true,
						Type:     types.NumberType,
					},
					"required_string": {
						Required: true,
						Type:     types.StringType,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
		},
	}, nil
}

func (rt testServeResourceTypeTwo) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
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
		"list_nested_blocks": tftypes.List{ElementType: tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"required_bool":   tftypes.Bool,
				"required_number": tftypes.Number,
				"required_string": tftypes.String,
			}},
		},
	},
}

type testServeResourceTwo struct {
	provider *testServeProvider
}

func (r testServeResourceTwo) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
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

func (r testServeResourceTwo) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	r.provider.readResourceCurrentStateValue = req.State.Raw
	r.provider.readResourceCurrentStateSchema = req.State.Schema
	r.provider.readResourceProviderMetaValue = req.ProviderMeta.Raw
	r.provider.readResourceProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.readResourceCalledResourceType = "test_two"
	r.provider.readResourceImpl(ctx, req, resp)
}

func (r testServeResourceTwo) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
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

func (r testServeResourceTwo) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	r.provider.applyResourceChangePriorStateValue = req.State.Raw
	r.provider.applyResourceChangePriorStateSchema = req.State.Schema
	r.provider.applyResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.applyResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.applyResourceChangeCalledResourceType = "test_two"
	r.provider.applyResourceChangeCalledAction = "delete"
	r.provider.deleteFunc(ctx, req, resp)
}

func (r testServeResourceTwo) ModifyPlan(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {
	r.provider.planResourceChangePriorStateValue = req.State.Raw
	r.provider.planResourceChangePriorStateSchema = req.State.Schema
	r.provider.planResourceChangeProposedNewStateValue = req.Plan.Raw
	r.provider.planResourceChangeProposedNewStateSchema = req.Plan.Schema
	r.provider.planResourceChangeConfigValue = req.Config.Raw
	r.provider.planResourceChangeConfigSchema = req.Config.Schema
	r.provider.planResourceChangeProviderMetaValue = req.ProviderMeta.Raw
	r.provider.planResourceChangeProviderMetaSchema = req.ProviderMeta.Schema
	r.provider.planResourceChangeCalledResourceType = "test_two"
	r.provider.planResourceChangeCalledAction = "modify_plan"
	if r.provider.modifyPlanFunc != nil {
		r.provider.modifyPlanFunc(ctx, req, resp)
	}
}
