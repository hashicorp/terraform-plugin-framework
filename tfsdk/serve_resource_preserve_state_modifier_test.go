package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func (rt testServeResourceTypePreserveStateModifier) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
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
				PlanModifiers: []AttributePlanModifier{
					PreserveState(),
				},
			},
		},
	}, nil
}

func (rt testServeResourceTypePreserveStateModifier) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServePreserveStateModifier{
		provider: provider,
	}, nil
}

var testServeResourceTypePreserveStateModifierSchema = &tfprotov6.Schema{
	Version: 1,
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "name",
				Required: true,
				Type:     tftypes.String,
			},
			{
				Name:     "last_updated",
				Required: true,
				Type:     tftypes.String,
			},
			{
				Name:     "first_updated",
				Required: true,
				Type:     tftypes.String,
			},
		},
	},
}

var testServeResourceTypePreserveStateModifierType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"name":          tftypes.String,
		"last_updated":  tftypes.String,
		"first_updated": tftypes.String,
	},
}

type testServePreserveStateModifier struct {
	provider *testServeProvider
}

type testServeResourceTypePreserveStateModifier struct{}

func (r testServePreserveStateModifier) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServePreserveStateModifier) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServePreserveStateModifier) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServePreserveStateModifier) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServePreserveStateModifier) ImportState(ctx context.Context, req ImportResourceStateRequest, resp *ImportResourceStateResponse) {
	ResourceImportStateNotImplemented(ctx, "Not expected to be called during testing.", resp)
}
