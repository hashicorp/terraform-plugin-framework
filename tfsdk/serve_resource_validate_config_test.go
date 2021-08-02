package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeValidateConfig struct{}

func (dt testServeResourceTypeValidateConfig) GetSchema(_ context.Context) (Schema, []*tfprotov6.Diagnostic) {
	return Schema{
		Attributes: map[string]Attribute{
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (dt testServeResourceTypeValidateConfig) NewResource(_ context.Context, p Provider) (Resource, []*tfprotov6.Diagnostic) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceValidateConfig{
		provider: provider,
	}, nil
}

var testServeResourceTypeValidateConfigSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "string",
				Optional: true,
				Type:     tftypes.String,
			},
		},
	},
}

var testServeResourceTypeValidateConfigType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"string": tftypes.String,
	},
}

type testServeResourceValidateConfig struct {
	provider *testServeProvider
}

func (r testServeResourceValidateConfig) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
}
func (r testServeResourceValidateConfig) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
}
func (r testServeResourceValidateConfig) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
}
func (r testServeResourceValidateConfig) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
}

func (r testServeResourceValidateConfig) ValidateConfig(ctx context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {
	r.provider.validateResourceConfigCalledResourceType = "test_validate_config"
	r.provider.validateResourceConfigImpl(ctx, req, resp)
}
