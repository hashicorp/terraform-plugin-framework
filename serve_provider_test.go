package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type testServeProvider struct {
	// configure
	configuredVal       tftypes.Value
	configuredSchema    schema.Schema
	configuredTFVersion string

	// read resource request
	readResourceCurrentStateValue  tftypes.Value
	readResourceCurrentStateSchema schema.Schema
	readResourceProviderMetaValue  tftypes.Value
	readResourceProviderMetaSchema schema.Schema
	readResourceImpl               func(context.Context, ReadResourceRequest, *ReadResourceResponse)
	readResourceCalledResourceType string
}

func (t *testServeProvider) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Version:            1,
		DeprecationMessage: "Deprecated in favor of other_resource",
		Attributes: map[string]schema.Attribute{
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
			// TODO: add maps when we support them
			// TODO: add sets when we support them
			// TODO: add tuples when we support them
			"single-nested-attributes": {
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
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
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}, schema.ListNestedAttributesOptions{}),
				Optional: true,
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
			// TODO: add maps when we support them
			// TODO: add sets when we support them
			// TODO: add tuples when we support them
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
		"list-string":       tftypes.List{ElementType: tftypes.String},
		"list-list-string":  tftypes.List{ElementType: tftypes.List{ElementType: tftypes.String}},
		"list-object": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Bool,
			"baz": tftypes.Number,
		}}},
		"object": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo":  tftypes.String,
			"bar":  tftypes.Bool,
			"baz":  tftypes.Number,
			"quux": tftypes.List{ElementType: tftypes.String},
		}},
		"empty-object": tftypes.Object{AttributeTypes: map[string]tftypes.Type{}},
		"single-nested-attributes": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
		}},
		"list-nested-attributes": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
		}}},
	},
}

func (t *testServeProvider) GetResources(_ context.Context) (map[string]ResourceType, []*tfprotov6.Diagnostic) {
	return map[string]ResourceType{
		"test_one": testServeResourceTypeOne{},
		"test_two": testServeResourceTwo{},
	}, nil
}

func (t *testServeProvider) GetDataSources(_ context.Context) (map[string]DataSourceType, []*tfprotov6.Diagnostic) {
	return map[string]DataSourceType{
		"test_one": testServeDataSourceOne{},
		"test_two": testServeDataSourceTwo{},
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

func (t *testServeProviderWithMetaSchema) GetMetaSchema(context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Version: 2,
		Attributes: map[string]schema.Attribute{
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
