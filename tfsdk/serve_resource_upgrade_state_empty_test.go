package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ ResourceWithUpgradeState = testServeResourceUpgradeStateEmpty{}

type testServeResourceTypeUpgradeStateEmpty struct{}

func (t testServeResourceTypeUpgradeStateEmpty) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"optional_attribute": {
				Type:     types.StringType,
				Optional: true,
			},
			"required_attribute": {
				Type:     types.StringType,
				Required: true,
			},
		},
		Version: 1, // Something above 0
	}, nil
}

func (t testServeResourceTypeUpgradeStateEmpty) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceUpgradeStateEmpty{
		provider: provider,
	}, nil
}

var testServeResourceTypeUpgradeStateEmptySchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "id",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "optional_attribute",
				Optional: true,
				Type:     tftypes.String,
			},
			{
				Name:     "required_attribute",
				Required: true,
				Type:     tftypes.String,
			},
		},
	},
	Version: 1,
}

type testServeResourceUpgradeStateEmpty struct {
	provider *testServeProvider
}

func (r testServeResourceUpgradeStateEmpty) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeStateEmpty) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeStateEmpty) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeStateEmpty) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeStateEmpty) ImportState(ctx context.Context, req ImportResourceStateRequest, resp *ImportResourceStateResponse) {
	ResourceImportStateNotImplemented(ctx, "intentionally not implemented", resp)
}

func (r testServeResourceUpgradeStateEmpty) UpgradeState(ctx context.Context) map[int64]ResourceStateUpgrader {
	return nil
}
