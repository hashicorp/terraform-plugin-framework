package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeProvider struct {
	// validate provider config request
	validateProviderConfigImpl func(context.Context, tfsdk.ValidateProviderConfigRequest, *tfsdk.ValidateProviderConfigResponse)

	// configure
	configuredVal       tftypes.Value
	configuredSchema    tfsdk.Schema
	configuredTFVersion string

	// validate resource config request
	validateResourceConfigCalledResourceType string
	validateResourceConfigImpl               func(context.Context, tfsdk.ValidateResourceConfigRequest, *tfsdk.ValidateResourceConfigResponse)

	// upgrade resource state
	upgradeResourceStateCalledResourceType string

	// read resource request
	readResourceCurrentStateValue  tftypes.Value
	readResourceCurrentStateSchema tfsdk.Schema
	readResourceProviderMetaValue  tftypes.Value
	readResourceProviderMetaSchema tfsdk.Schema
	readResourceImpl               func(context.Context, tfsdk.ReadResourceRequest, *tfsdk.ReadResourceResponse)
	readResourceCalledResourceType string

	// plan resource change
	planResourceChangeCalledResourceType     string
	planResourceChangeCalledAction           string
	planResourceChangePriorStateValue        tftypes.Value
	planResourceChangePriorStateSchema       tfsdk.Schema
	planResourceChangeProposedNewStateValue  tftypes.Value
	planResourceChangeProposedNewStateSchema tfsdk.Schema
	planResourceChangeConfigValue            tftypes.Value
	planResourceChangeConfigSchema           tfsdk.Schema
	planResourceChangeProviderMetaValue      tftypes.Value
	planResourceChangeProviderMetaSchema     tfsdk.Schema
	modifyPlanFunc                           func(context.Context, tfsdk.ModifyResourcePlanRequest, *tfsdk.ModifyResourcePlanResponse)

	// apply resource change
	applyResourceChangeCalledResourceType string
	applyResourceChangeCalledAction       string
	applyResourceChangePriorStateValue    tftypes.Value
	applyResourceChangePriorStateSchema   tfsdk.Schema
	applyResourceChangePlannedStateValue  tftypes.Value
	applyResourceChangePlannedStateSchema tfsdk.Schema
	applyResourceChangeConfigValue        tftypes.Value
	applyResourceChangeConfigSchema       tfsdk.Schema
	applyResourceChangeProviderMetaValue  tftypes.Value
	applyResourceChangeProviderMetaSchema tfsdk.Schema
	createFunc                            func(context.Context, tfsdk.CreateResourceRequest, *tfsdk.CreateResourceResponse)
	updateFunc                            func(context.Context, tfsdk.UpdateResourceRequest, *tfsdk.UpdateResourceResponse)
	deleteFunc                            func(context.Context, tfsdk.DeleteResourceRequest, *tfsdk.DeleteResourceResponse)

	// import resource state
	importResourceStateCalledResourceType string
	importStateFunc                       func(context.Context, tfsdk.ImportResourceStateRequest, *tfsdk.ImportResourceStateResponse)

	// validate data source config request
	validateDataSourceConfigCalledDataSourceType string
	validateDataSourceConfigImpl                 func(context.Context, tfsdk.ValidateDataSourceConfigRequest, *tfsdk.ValidateDataSourceConfigResponse)

	// read data source request
	readDataSourceConfigValue          tftypes.Value
	readDataSourceConfigSchema         tfsdk.Schema
	readDataSourceProviderMetaValue    tftypes.Value
	readDataSourceProviderMetaSchema   tfsdk.Schema
	readDataSourceImpl                 func(context.Context, tfsdk.ReadDataSourceRequest, *tfsdk.ReadDataSourceResponse)
	readDataSourceCalledDataSourceType string
}

func (t *testServeProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Version:            1,
		DeprecationMessage: "Deprecated in favor of other_resource",
		Attributes: map[string]tfsdk.Attribute{
			"required": {
				Type:     types.StringType,
				Required: true,
			},
			"optional": {
				Type:     types.StringType,
				Optional: true,
			},
			"computed": {
				Type:     types.StringType,
				Computed: true,
			},
			"optional_computed": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"sensitive": {
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
			"deprecated": {
				Type:               types.StringType,
				Optional:           true,
				DeprecationMessage: "Deprecated, please use \"optional\" instead",
			},
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
			"number": {
				Type:     types.NumberType,
				Optional: true,
			},
			"bool": {
				Type:     types.BoolType,
				Optional: true,
			},
			"int64": {
				Type:     types.Int64Type,
				Optional: true,
			},
			"float64": {
				Type:     types.Float64Type,
				Optional: true,
			},
			"list-string": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"list-list-string": {
				Type: types.ListType{
					ElemType: types.ListType{
						ElemType: types.StringType,
					},
				},
				Optional: true,
			},
			"list-object": {
				Type: types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"foo": types.StringType,
							"bar": types.BoolType,
							"baz": types.NumberType,
						},
					},
				},
				Optional: true,
			},
			"object": {
				Type: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"foo": types.StringType,
						"bar": types.BoolType,
						"baz": types.NumberType,
						"quux": types.ListType{
							ElemType: types.StringType,
						},
					},
				},
				Optional: true,
			},
			"empty-object": {
				Type:     types.ObjectType{},
				Optional: true,
			},
			"map": {
				Type:     types.MapType{ElemType: types.NumberType},
				Optional: true,
			},
			"set-string": {
				Type: types.SetType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"set-set-string": {
				Type: types.SetType{
					ElemType: types.SetType{
						ElemType: types.StringType,
					},
				},
				Optional: true,
			},
			"set-object": {
				Type: types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"foo": types.StringType,
							"bar": types.BoolType,
							"baz": types.NumberType,
						},
					},
				},
				Optional: true,
			},
			// TODO: add tuples when we support them
			"single-nested-attributes": {
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}),
				Optional: true,
			},
			"list-nested-attributes": {
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}),
				Optional: true,
			},
			"map-nested-attributes": {
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}),
				Optional: true,
			},
			"set-nested-attributes": {
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}),
				Optional: true,
			},
		},
		Blocks: map[string]tfsdk.Block{
			"list-nested-blocks": {
				Attributes: map[string]tfsdk.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
			"set-nested-blocks": {
				Attributes: map[string]tfsdk.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
			},
		},
	}, nil
}

func (t *testServeProvider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"test_one":                          testServeResourceTypeOne{},
		"test_two":                          testServeResourceTypeTwo{},
		"test_three":                        testServeResourceTypeThree{},
		"test_attribute_plan_modifiers":     testServeResourceTypeAttributePlanModifiers{},
		"test_import_state":                 testServeResourceTypeImportState{},
		"test_import_state_not_implemented": testServeResourceTypeImportStateNotImplemented{},
	}, nil
}

func (t *testServeProvider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"test_one": testServeDataSourceTypeOne{},
		"test_two": testServeDataSourceTypeTwo{},
	}, nil
}

func (t *testServeProvider) Configure(_ context.Context, req tfsdk.ConfigureProviderRequest, _ *tfsdk.ConfigureProviderResponse) {
	t.configuredVal = req.Config.Raw
	t.configuredSchema = req.Config.Schema
	t.configuredTFVersion = req.TerraformVersion
}

type testServeProviderWithMetaSchema struct {
	*testServeProvider
}

func (t *testServeProviderWithMetaSchema) GetMetaSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Version: 2,
		Attributes: map[string]tfsdk.Attribute{
			"foo": {
				Type:                types.StringType,
				Required:            true,
				Description:         "A string",
				MarkdownDescription: "A **string**",
			},
		},
	}, nil
}

var testServeProviderMetaType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"foo": tftypes.String,
	},
}
