package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeImportState struct{}

func (dt testServeResourceTypeImportState) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
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

func (dt testServeResourceTypeImportState) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceImportState{
		provider: provider,
	}, nil
}

var testServeResourceTypeImportStateSchema = &tfprotov6.Schema{
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

var testServeResourceTypeImportStateTftype = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":              tftypes.String,
		"optional_string": tftypes.String,
		"required_string": tftypes.String,
	},
}

type testServeResourceImportStateData struct {
	Id             string  `tfsdk:"id"`
	OptionalString *string `tfsdk:"optional_string"`
	RequiredString string  `tfsdk:"required_string"`
}

type testServeResourceImportState struct {
	provider *testServeProvider
}

func (r testServeResourceImportState) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceImportState) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceImportState) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceImportState) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceImportState) ImportState(ctx context.Context, req ImportResourceStateRequest, resp *ImportResourceStateResponse) {
	r.provider.importResourceStateCalledResourceType = "test_import_state"
	r.provider.importStateFunc(ctx, req, resp)
}
