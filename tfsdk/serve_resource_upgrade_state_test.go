package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// This resource is a placeholder for UpgradeResourceState testing,
// so it is decoupled from other test resources.
// TODO: Implement UpgradeResourceState support, when added
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/42
type testServeResourceTypeUpgradeState struct{}

func (t testServeResourceTypeUpgradeState) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"optional_string": {
				Type:     types.StringType,
				Optional: true,
			},
			"required_string": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (t testServeResourceTypeUpgradeState) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceUpgradeState{
		provider: provider,
	}, nil
}

var testServeResourceTypeUpgradeStateSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "id",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "optional_string",
				Optional: true,
				Type:     tftypes.String,
			},
			{
				Name:     "required_string",
				Required: true,
				Type:     tftypes.String,
			},
		},
	},
}

var testServeResourceTypeUpgradeStateTftype = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":              tftypes.String,
		"optional_string": tftypes.String,
		"required_string": tftypes.String,
	},
}

type testServeResourceUpgradeStateData struct {
	Id             string  `tfsdk:"id"`
	OptionalString *string `tfsdk:"optional_string"`
	RequiredString string  `tfsdk:"required_string"`
}

type testServeResourceUpgradeState struct {
	provider *testServeProvider
}

func (r testServeResourceUpgradeState) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeState) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeState) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeState) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeState) ImportState(ctx context.Context, req ImportResourceStateRequest, resp *ImportResourceStateResponse) {
	ResourceImportStateNotImplemented(ctx, "intentionally not implemented", resp)
}
