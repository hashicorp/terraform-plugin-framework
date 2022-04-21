package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeImportStateNotImplemented struct{}

func (dt testServeResourceTypeImportStateNotImplemented) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (dt testServeResourceTypeImportStateNotImplemented) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceImportStateNotImplemented{
		provider: provider,
	}, nil
}

var testServeResourceTypeImportStateNotImplementedSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "id",
				Computed: true,
				Type:     tftypes.String,
			},
		},
	},
}

type testServeResourceImportStateNotImplemented struct {
	provider *testServeProvider
}

func (r testServeResourceImportStateNotImplemented) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceImportStateNotImplemented) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceImportStateNotImplemented) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceImportStateNotImplemented) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
