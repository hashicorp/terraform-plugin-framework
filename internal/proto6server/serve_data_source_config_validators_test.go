package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeDataSourceTypeConfigValidators struct{}

func (dt testServeDataSourceTypeConfigValidators) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (dt testServeDataSourceTypeConfigValidators) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, ok := p.(*testServeProvider)
	if !ok {
		prov, ok := p.(*testServeProviderWithMetaSchema)
		if !ok {
			panic(fmt.Sprintf("unexpected provider type %T", p))
		}
		provider = prov.testServeProvider
	}
	return testServeDataSourceConfigValidators{
		provider: provider,
	}, nil
}

var testServeDataSourceTypeConfigValidatorsSchema = &tfprotov6.Schema{
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

var testServeDataSourceTypeConfigValidatorsType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"string": tftypes.String,
	},
}

type testServeDataSourceConfigValidators struct {
	provider *testServeProvider
}

func (r testServeDataSourceConfigValidators) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	// Intentionally blank. Not expected to be called during testing.
}

func (r testServeDataSourceConfigValidators) ConfigValidators(ctx context.Context) []tfsdk.DataSourceConfigValidator {
	r.provider.validateDataSourceConfigCalledDataSourceType = "test_config_validators"

	return []tfsdk.DataSourceConfigValidator{
		newTestDataSourceConfigValidator(r.provider.validateDataSourceConfigImpl),
		// Verify multiple validators
		newTestDataSourceConfigValidator(r.provider.validateDataSourceConfigImpl),
	}
}

type testDataSourceConfigValidator struct {
	tfsdk.DataSourceConfigValidator

	impl func(context.Context, tfsdk.ValidateDataSourceConfigRequest, *tfsdk.ValidateDataSourceConfigResponse)
}

func (v testDataSourceConfigValidator) Description(ctx context.Context) string {
	return "test data source config validator"
}
func (v testDataSourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "**test** data source config validator"
}
func (v testDataSourceConfigValidator) Validate(ctx context.Context, req tfsdk.ValidateDataSourceConfigRequest, resp *tfsdk.ValidateDataSourceConfigResponse) {
	v.impl(ctx, req, resp)
}

func newTestDataSourceConfigValidator(impl func(context.Context, tfsdk.ValidateDataSourceConfigRequest, *tfsdk.ValidateDataSourceConfigResponse)) testDataSourceConfigValidator {
	return testDataSourceConfigValidator{impl: impl}
}
