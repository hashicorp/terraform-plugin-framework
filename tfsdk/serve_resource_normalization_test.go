package tfsdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeNormalization struct{}

func (dt testServeResourceTypeNormalization) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     testServeResourceTypeNormalizationStringType,
				Required: true,
			},
		},
	}, nil
}

func (dt testServeResourceTypeNormalization) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceNormalization{
		provider: provider,
	}, nil
}

var testServeResourceTypeNormalizationSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "id",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "name",
				Required: true,
				Type:     tftypes.String,
			},
		},
	},
}

var testServeResourceTypeNormalizationStringType = types.SpecializedStringType(types.SpecializedStringOpts{
	TypeString: "tfsdk.testServeResourceTypeNormalizationStringType",
	NormalizeFunc: func(given string) string {
		return strings.ToLower(given)
	},
})

var testServeResourceTypeNormalizationType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	},
}

type testServeResourceNormalization struct {
	provider *testServeProvider
}

func (r testServeResourceNormalization) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	r.provider.applyResourceChangeCalledResourceType = "test_normalization"
	r.provider.applyResourceChangeCalledAction = "create"
}
func (r testServeResourceNormalization) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	r.provider.readResourceCalledResourceType = "test_normalization"
	r.provider.readResourceCurrentStateValue = req.State.Raw
	r.provider.readResourceCurrentStateSchema = req.State.Schema
	r.provider.readResourceImpl(ctx, req, resp)
}
func (r testServeResourceNormalization) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	r.provider.applyResourceChangeCalledResourceType = "test_normalization"
	r.provider.applyResourceChangeCalledAction = "update"
}
func (r testServeResourceNormalization) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	r.provider.applyResourceChangeCalledResourceType = "test_normalization"
	r.provider.applyResourceChangeCalledAction = "delete"
}
func (r testServeResourceNormalization) ImportState(ctx context.Context, req ImportResourceStateRequest, resp *ImportResourceStateResponse) {
	ResourceImportStateNotImplemented(ctx, "Not expected to be called during testing.", resp)
}

func (r testServeResourceNormalization) ValidateConfig(ctx context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {
	r.provider.validateResourceConfigCalledResourceType = "test_normalization"
	r.provider.validateResourceConfigImpl(ctx, req, resp)
}
