package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func (rt testServeResourceTypeThree) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Version: 1,
		Attributes: map[string]tfsdk.Attribute{
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
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"computed_string": {
						Computed: true,
						Type:     types.StringType,
					},
					"string": {
						Optional: true,
						Type:     types.StringType,
					},
				}, tfsdk.MapNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (rt testServeResourceTypeThree) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
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

func (r testServeResourceThree) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceThree) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceThree) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeResourceThree) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
