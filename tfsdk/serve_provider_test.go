package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeProvider struct {
	// validate provider config request
	validateProviderConfigImpl func(context.Context, ValidateProviderConfigRequest, *ValidateProviderConfigResponse)

	// configure
	configuredVal       tftypes.Value
	configuredSchema    Schema
	configuredTFVersion string

	// validate resource config request
	validateResourceConfigCalledResourceType string
	validateResourceConfigImpl               func(context.Context, ValidateResourceConfigRequest, *ValidateResourceConfigResponse)

	// read resource request
	readResourceCurrentStateValue  tftypes.Value
	readResourceCurrentStateSchema Schema
	readResourceProviderMetaValue  tftypes.Value
	readResourceProviderMetaSchema Schema
	readResourceImpl               func(context.Context, ReadResourceRequest, *ReadResourceResponse)
	readResourceCalledResourceType string

	// plan resource change
	planResourceChangeCalledResourceType     string
	planResourceChangeCalledAction           string
	planResourceChangePriorStateValue        tftypes.Value
	planResourceChangePriorStateSchema       Schema
	planResourceChangeProposedNewStateValue  tftypes.Value
	planResourceChangeProposedNewStateSchema Schema
	planResourceChangeConfigValue            tftypes.Value
	planResourceChangeConfigSchema           Schema
	planResourceChangeProviderMetaValue      tftypes.Value
	planResourceChangeProviderMetaSchema     Schema
	modifyPlanFunc                           func(context.Context, ModifyResourcePlanRequest, *ModifyResourcePlanResponse)

	// apply resource change
	applyResourceChangeCalledResourceType string
	applyResourceChangeCalledAction       string
	applyResourceChangePriorStateValue    tftypes.Value
	applyResourceChangePriorStateSchema   Schema
	applyResourceChangePlannedStateValue  tftypes.Value
	applyResourceChangePlannedStateSchema Schema
	applyResourceChangeConfigValue        tftypes.Value
	applyResourceChangeConfigSchema       Schema
	applyResourceChangeProviderMetaValue  tftypes.Value
	applyResourceChangeProviderMetaSchema Schema
	createFunc                            func(context.Context, CreateResourceRequest, *CreateResourceResponse)
	updateFunc                            func(context.Context, UpdateResourceRequest, *UpdateResourceResponse)
	deleteFunc                            func(context.Context, DeleteResourceRequest, *DeleteResourceResponse)

	// import resource state
	importResourceStateCalledResourceType string
	importStateFunc                       func(context.Context, ImportResourceStateRequest, *ImportResourceStateResponse)

	// validate data source config request
	validateDataSourceConfigCalledDataSourceType string
	validateDataSourceConfigImpl                 func(context.Context, ValidateDataSourceConfigRequest, *ValidateDataSourceConfigResponse)

	// read data source request
	readDataSourceConfigValue          tftypes.Value
	readDataSourceConfigSchema         Schema
	readDataSourceProviderMetaValue    tftypes.Value
	readDataSourceProviderMetaSchema   Schema
	readDataSourceImpl                 func(context.Context, ReadDataSourceRequest, *ReadDataSourceResponse)
	readDataSourceCalledDataSourceType string
}

func (t *testServeProvider) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Version:            1,
		DeprecationMessage: "Deprecated in favor of other_resource",
		Attributes: map[string]Attribute{
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
				Attributes: SingleNestedAttributes(map[string]Attribute{
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
				Attributes: ListNestedAttributes(map[string]Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}, ListNestedAttributesOptions{}),
				Optional: true,
			},
			"map-nested-attributes": {
				Attributes: MapNestedAttributes(map[string]Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}, MapNestedAttributesOptions{}),
				Optional: true,
			},
			"set-nested-attributes": {
				Attributes: SetNestedAttributes(map[string]Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}, SetNestedAttributesOptions{}),
				Optional: true,
			},
		},
		Blocks: map[string]Block{
			"list-nested-blocks": {
				Attributes: map[string]Attribute{
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
				NestingMode: BlockNestingModeList,
			},
			"set-nested-blocks": {
				Attributes: map[string]Attribute{
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
				NestingMode: BlockNestingModeSet,
			},
		},
	}, nil
}

var testServeProviderProviderSchema = &tfprotov6.Schema{
	Version: 1,
	Block: &tfprotov6.SchemaBlock{
		Deprecated: true,
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "bool",
				Type:     tftypes.Bool,
				Optional: true,
			},
			{
				Name:     "computed",
				Type:     tftypes.String,
				Computed: true,
			},
			{
				Name:       "deprecated",
				Type:       tftypes.String,
				Optional:   true,
				Deprecated: true,
			},
			{
				Name: "empty-object",
				Type: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{},
				},
				Optional: true,
			},
			{
				Name:     "float64",
				Type:     tftypes.Number,
				Optional: true,
			},
			{
				Name:     "int64",
				Type:     tftypes.Number,
				Optional: true,
			},
			{
				Name: "list-list-string",
				Type: tftypes.List{
					ElementType: tftypes.List{
						ElementType: tftypes.String,
					},
				},
				Optional: true,
			},
			{
				Name: "list-nested-attributes",
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bar",
							Type:     tftypes.Number,
							Required: true,
						},
						{
							Name:     "foo",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
			},
			{
				Name: "list-object",
				Type: tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"foo": tftypes.String,
							"bar": tftypes.Bool,
							"baz": tftypes.Number,
						},
					},
				},
				Optional: true,
			},
			{
				Name: "list-string",
				Type: tftypes.List{
					ElementType: tftypes.String,
				},
				Optional: true,
			},
			{
				Name: "map",
				Type: tftypes.Map{
					ElementType: tftypes.Number,
				},
				Optional: true,
			},
			{
				Name:     "map-nested-attributes",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeMap,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bar",
							Type:     tftypes.Number,
							Required: true,
						},
						{
							Name:     "foo",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			{
				Name:     "number",
				Type:     tftypes.Number,
				Optional: true,
			},
			{
				Name: "object",
				Type: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Bool,
						"baz": tftypes.Number,
						"quux": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				Optional: true,
			},
			{
				Name:     "optional",
				Type:     tftypes.String,
				Optional: true,
			},
			{
				Name:     "optional_computed",
				Type:     tftypes.String,
				Optional: true,
				Computed: true,
			},
			{
				Name:     "required",
				Type:     tftypes.String,
				Required: true,
			},
			{
				Name:      "sensitive",
				Type:      tftypes.String,
				Optional:  true,
				Sensitive: true,
			},
			{
				Name: "set-nested-attributes",
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bar",
							Type:     tftypes.Number,
							Required: true,
						},
						{
							Name:     "foo",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
			},
			{
				Name: "set-object",
				Type: tftypes.Set{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"foo": tftypes.String,
							"bar": tftypes.Bool,
							"baz": tftypes.Number,
						},
					},
				},
				Optional: true,
			},
			{
				Name: "set-set-string",
				Type: tftypes.Set{
					ElementType: tftypes.Set{
						ElementType: tftypes.String,
					},
				},
				Optional: true,
			},
			{
				Name: "set-string",
				Type: tftypes.Set{
					ElementType: tftypes.String,
				},
				Optional: true,
			},
			{
				Name: "single-nested-attributes",
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSingle,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bar",
							Type:     tftypes.Number,
							Required: true,
						},
						{
							Name:     "foo",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
			},
			{
				Name:     "string",
				Type:     tftypes.String,
				Optional: true,
			},
			// TODO: add tuples when we support them
		},
		BlockTypes: []*tfprotov6.SchemaNestedBlock{
			{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bar",
							Type:     tftypes.Number,
							Required: true,
						},
						{
							Name:     "foo",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "list-nested-blocks",
			},
			{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bar",
							Type:     tftypes.Number,
							Required: true,
						},
						{
							Name:     "foo",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
				TypeName: "set-nested-blocks",
			},
		},
	},
}

var testServeProviderProviderType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"required":          tftypes.String,
		"optional":          tftypes.String,
		"computed":          tftypes.String,
		"optional_computed": tftypes.String,
		"sensitive":         tftypes.String,
		"deprecated":        tftypes.String,
		"string":            tftypes.String,
		"number":            tftypes.Number,
		"bool":              tftypes.Bool,
		"int64":             tftypes.Number,
		"float64":           tftypes.Number,
		"list-string":       tftypes.List{ElementType: tftypes.String},
		"list-list-string":  tftypes.List{ElementType: tftypes.List{ElementType: tftypes.String}},
		"list-object": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Bool,
			"baz": tftypes.Number,
		}}},
		"map": tftypes.Map{ElementType: tftypes.Number},
		"object": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo":  tftypes.String,
			"bar":  tftypes.Bool,
			"baz":  tftypes.Number,
			"quux": tftypes.List{ElementType: tftypes.String},
		}},
		"set-string":     tftypes.Set{ElementType: tftypes.String},
		"set-set-string": tftypes.Set{ElementType: tftypes.Set{ElementType: tftypes.String}},
		"set-object": tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Bool,
			"baz": tftypes.Number,
		}}},
		"empty-object": tftypes.Object{AttributeTypes: map[string]tftypes.Type{}},
		"single-nested-attributes": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
		}},
		"list-nested-attributes": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
		}}},
		"list-nested-blocks": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
		}}},
		"map-nested-attributes": tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
		}}},
		"set-nested-attributes": tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
		}}},
		"set-nested-blocks": tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
		}}},
	},
}

func (t *testServeProvider) GetResources(_ context.Context) (map[string]ResourceType, diag.Diagnostics) {
	return map[string]ResourceType{
		"test_one":                      testServeResourceTypeOne{},
		"test_two":                      testServeResourceTypeTwo{},
		"test_three":                    testServeResourceTypeThree{},
		"test_attribute_plan_modifiers": testServeResourceTypeAttributePlanModifiers{},
		"test_config_validators":        testServeResourceTypeConfigValidators{},
		"test_import_state":             testServeResourceTypeImportState{},
		"test_validate_config":          testServeResourceTypeValidateConfig{},
	}, nil
}

func (t *testServeProvider) GetDataSources(_ context.Context) (map[string]DataSourceType, diag.Diagnostics) {
	return map[string]DataSourceType{
		"test_one":               testServeDataSourceTypeOne{},
		"test_two":               testServeDataSourceTypeTwo{},
		"test_config_validators": testServeDataSourceTypeConfigValidators{},
		"test_validate_config":   testServeDataSourceTypeValidateConfig{},
	}, nil
}

func (t *testServeProvider) Configure(_ context.Context, req ConfigureProviderRequest, _ *ConfigureProviderResponse) {
	t.configuredVal = req.Config.Raw
	t.configuredSchema = req.Config.Schema
	t.configuredTFVersion = req.TerraformVersion
}

type testServeProviderWithMetaSchema struct {
	*testServeProvider
}

func (t *testServeProviderWithMetaSchema) GetMetaSchema(context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Version: 2,
		Attributes: map[string]Attribute{
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
