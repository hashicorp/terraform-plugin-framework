package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeResourceTypeConfigValidators struct{}

func (dt testServeResourceTypeConfigValidators) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (dt testServeResourceTypeConfigValidators) NewResource(_ context.Context, p Provider) (Resource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeResourceConfigValidators{
		provider: provider,
	}, nil
}

var testServeResourceTypeConfigValidatorsSchema = &tfprotov6.Schema{
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

var testServeResourceTypeConfigValidatorsType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"string": tftypes.String,
	},
}

type testServeResourceConfigValidators struct {
	provider *testServeProvider
}

func (r testServeResourceConfigValidators) Create(ctx context.Context, req CreateResourceRequest, resp *CreateResourceResponse) {
}
func (r testServeResourceConfigValidators) Read(ctx context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
}
func (r testServeResourceConfigValidators) Update(ctx context.Context, req UpdateResourceRequest, resp *UpdateResourceResponse) {
}
func (r testServeResourceConfigValidators) Delete(ctx context.Context, req DeleteResourceRequest, resp *DeleteResourceResponse) {
}

func (r testServeResourceConfigValidators) ConfigValidators(ctx context.Context) []ResourceConfigValidator {
	r.provider.validateResourceConfigCalledResourceType = "test_config_validators"

	return []ResourceConfigValidator{
		newTestResourceConfigValidator(r.provider.validateResourceConfigImpl),
		// Verify multiple validators
		newTestResourceConfigValidator(r.provider.validateResourceConfigImpl),
	}
}

type testResourceConfigValidator struct {
	ResourceConfigValidator

	impl func(context.Context, ValidateResourceConfigRequest, *ValidateResourceConfigResponse)
}

func (v testResourceConfigValidator) Description(ctx context.Context) string {
	return "test resource config validator"
}
func (v testResourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "**test** resource config validator"
}
func (v testResourceConfigValidator) Validate(ctx context.Context, req ValidateResourceConfigRequest, resp *ValidateResourceConfigResponse) {
	v.impl(ctx, req, resp)
}

func newTestResourceConfigValidator(impl func(context.Context, ValidateResourceConfigRequest, *ValidateResourceConfigResponse)) testResourceConfigValidator {
	return testResourceConfigValidator{impl: impl}
}
