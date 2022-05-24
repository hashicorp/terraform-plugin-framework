package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type testServeResourceTypeUpgradeStateNotImplemented struct{}

func (t testServeResourceTypeUpgradeStateNotImplemented) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
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

func (t testServeResourceTypeUpgradeStateNotImplemented) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceUpgradeStateNotImplemented{
		provider: provider,
	}, nil
}

type testServeResourceUpgradeStateNotImplemented struct {
	provider *testServeProvider
}

func (r testServeResourceUpgradeStateNotImplemented) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeStateNotImplemented) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeStateNotImplemented) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
func (r testServeResourceUpgradeStateNotImplemented) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}
