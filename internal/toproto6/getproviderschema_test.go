package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TODO: DynamicPseudoType support
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/147
// TODO: Tuple type support
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/54
func TestGetProviderSchemaResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    *fwserver.GetProviderSchemaResponse
		expected *tfprotov6.GetProviderSchemaResponse
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"data-source-multiple-data-sources": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source_1": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed: true,
								Type:     types.BoolType,
							},
						},
					},
					"test_data_source_2": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source_1": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed: true,
									Name:     "test_attribute",
									Type:     tftypes.Bool,
								},
							},
						},
					},
					"test_data_source_2": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed: true,
									Name:     "test_attribute",
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed: true,
									Name:     "test_attribute",
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-deprecated": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								DeprecationMessage: "deprecated",
								Optional:           true,
								Type:               types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Deprecated: true,
									Name:       "test_attribute",
									Optional:   true,
									Type:       tftypes.Bool,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-optional": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Optional: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Optional: true,
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-optional-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed: true,
								Optional: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed: true,
									Name:     "test_attribute",
									Optional: true,
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Type:     tftypes.Bool,
									Required: true,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-sensitive": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed:  true,
								Sensitive: true,
								Type:      types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed:  true,
									Name:      "test_attribute",
									Sensitive: true,
									Type:      tftypes.Bool,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-bool": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-float64": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.Float64Type,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.Number,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-int64": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.Int64Type,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.Number,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-list-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.ListType{
									ElemType: types.ListType{
										ElemType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.List{
										ElementType: tftypes.List{
											ElementType: tftypes.String,
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-list-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"test_nested_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name: "test_attribute",
									NestedType: &tfprotov6.SchemaObject{
										Nesting: tfprotov6.SchemaObjectNestingModeList,
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_nested_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-list-object": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"test_object_attribute": types.StringType,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.List{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_object_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.List{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-map-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
									"test_nested_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name: "test_attribute",
									NestedType: &tfprotov6.SchemaObject{
										Nesting: tfprotov6.SchemaObjectNestingModeMap,
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_nested_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-map-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.MapType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Map{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-number": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.NumberType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.Number,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-object": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"test_object_attribute": types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_object_attribute": tftypes.String,
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-set-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"test_nested_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name: "test_attribute",
									NestedType: &tfprotov6.SchemaObject{
										Nesting: tfprotov6.SchemaObjectNestingModeSet,
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_nested_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-set-object": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.SetType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"test_object_attribute": types.StringType,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Set{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_object_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-set-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.SetType{
									ElemType: types.SetType{
										ElemType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Set{
										ElementType: tftypes.Set{
											ElementType: tftypes.String,
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Set{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-single-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"test_nested_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name: "test_attribute",
									NestedType: &tfprotov6.SchemaObject{
										Nesting: tfprotov6.SchemaObjectNestingModeSingle,
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_nested_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-block-list": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test_block": {
								Attributes: map[string]tfsdk.Attribute{
									"test_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							BlockTypes: []*tfprotov6.SchemaNestedBlock{
								{
									Block: &tfprotov6.SchemaBlock{
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
									TypeName: "test_block",
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-block-set": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test_block": {
								Attributes: map[string]tfsdk.Attribute{
									"test_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							BlockTypes: []*tfprotov6.SchemaNestedBlock{
								{
									Block: &tfprotov6.SchemaBlock{
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
									TypeName: "test_block",
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-block-single": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test_block": {
								Attributes: map[string]tfsdk.Attribute{
									"test_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							BlockTypes: []*tfprotov6.SchemaNestedBlock{
								{
									Block: &tfprotov6.SchemaBlock{
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
									TypeName: "test_block",
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-version": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": &tfsdk.Schema{
						Version: 123,
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block:   &tfprotov6.SchemaBlock{},
						Version: 123,
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Computed: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Computed: true,
								Name:     "test_attribute",
								Type:     tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-deprecated": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							DeprecationMessage: "deprecated",
							Optional:           true,
							Type:               types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Deprecated: true,
								Name:       "test_attribute",
								Optional:   true,
								Type:       tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-optional": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Optional: true,
								Type:     tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-optional-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Computed: true,
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Computed: true,
								Name:     "test_attribute",
								Optional: true,
								Type:     tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Type:     tftypes.Bool,
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-sensitive": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Computed:  true,
							Sensitive: true,
							Type:      types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Computed:  true,
								Name:      "test_attribute",
								Sensitive: true,
								Type:      tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-bool": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-float64": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.Float64Type,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.Number,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-int64": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.Int64Type,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.Number,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-list-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.ListType{
								ElemType: types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.List{
									ElementType: tftypes.List{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-list-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"test_nested_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name: "test_attribute",
								NestedType: &tfprotov6.SchemaObject{
									Nesting: tfprotov6.SchemaObjectNestingModeList,
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_nested_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-list-object": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"test_object_attribute": types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_object_attribute": tftypes.String,
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.List{
									ElementType: tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-map-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"test_nested_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name: "test_attribute",
								NestedType: &tfprotov6.SchemaObject{
									Nesting: tfprotov6.SchemaObjectNestingModeMap,
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_nested_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-map-string": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Map{
									ElementType: tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-number": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.NumberType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.Number,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-object": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"test_object_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_object_attribute": tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-set-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"test_nested_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name: "test_attribute",
								NestedType: &tfprotov6.SchemaObject{
									Nesting: tfprotov6.SchemaObjectNestingModeSet,
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_nested_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-set-object": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.SetType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"test_object_attribute": types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_object_attribute": tftypes.String,
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-set-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.SetType{
								ElemType: types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Set{
									ElementType: tftypes.Set{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Set{
									ElementType: tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-single-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"test_nested_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name: "test_attribute",
								NestedType: &tfprotov6.SchemaObject{
									Nesting: tfprotov6.SchemaObjectNestingModeSingle,
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_nested_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-type-string": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.StringType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.String,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-block-list": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"test_block": {
							Attributes: map[string]tfsdk.Attribute{
								"test_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						BlockTypes: []*tfprotov6.SchemaNestedBlock{
							{
								Block: &tfprotov6.SchemaBlock{
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
								TypeName: "test_block",
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-block-set": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"test_block": {
							Attributes: map[string]tfsdk.Attribute{
								"test_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						BlockTypes: []*tfprotov6.SchemaNestedBlock{
							{
								Block: &tfprotov6.SchemaBlock{
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
								TypeName: "test_block",
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-block-single": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"test_block": {
							Attributes: map[string]tfsdk.Attribute{
								"test_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						BlockTypes: []*tfprotov6.SchemaNestedBlock{
							{
								Block: &tfprotov6.SchemaBlock{
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
								TypeName: "test_block",
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-version": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: &tfsdk.Schema{
					Version: 123,
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block:   &tfprotov6.SchemaBlock{},
					Version: 123,
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Computed: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Computed: true,
								Name:     "test_attribute",
								Type:     tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-deprecated": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							DeprecationMessage: "deprecated",
							Optional:           true,
							Type:               types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Deprecated: true,
								Name:       "test_attribute",
								Optional:   true,
								Type:       tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-optional": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Optional: true,
								Type:     tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-optional-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Computed: true,
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Computed: true,
								Name:     "test_attribute",
								Optional: true,
								Type:     tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Type:     tftypes.Bool,
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-sensitive": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Computed:  true,
							Sensitive: true,
							Type:      types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Computed:  true,
								Name:      "test_attribute",
								Sensitive: true,
								Type:      tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-bool": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-float64": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.Float64Type,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.Number,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-int64": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.Int64Type,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.Number,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-list-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.ListType{
								ElemType: types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.List{
									ElementType: tftypes.List{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-list-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"test_nested_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name: "test_attribute",
								NestedType: &tfprotov6.SchemaObject{
									Nesting: tfprotov6.SchemaObjectNestingModeList,
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_nested_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-list-object": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.ListType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"test_nested_attribute": types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_nested_attribute": tftypes.String,
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.List{
									ElementType: tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-map-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"test_nested_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name: "test_attribute",
								NestedType: &tfprotov6.SchemaObject{
									Nesting: tfprotov6.SchemaObjectNestingModeMap,
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_nested_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-map-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.MapType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Map{
									ElementType: tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-number": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.NumberType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.Number,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-object": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"test_object_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_object_attribute": tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-set-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"test_nested_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name: "test_attribute",
								NestedType: &tfprotov6.SchemaObject{
									Nesting: tfprotov6.SchemaObjectNestingModeSet,
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_nested_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-set-object": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.SetType{
								ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"test_object_attribute": types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_object_attribute": tftypes.String,
										},
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-set-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.SetType{
								ElemType: types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Set{
									ElementType: tftypes.Set{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type: tftypes.Set{
									ElementType: tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-single-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"test_nested_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name: "test_attribute",
								NestedType: &tfprotov6.SchemaObject{
									Nesting: tfprotov6.SchemaObjectNestingModeSingle,
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_nested_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-attribute-type-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test_attribute": {
							Required: true,
							Type:     types.StringType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.String,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-block-list": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"test_block": {
							Attributes: map[string]tfsdk.Attribute{
								"test_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						BlockTypes: []*tfprotov6.SchemaNestedBlock{
							{
								Block: &tfprotov6.SchemaBlock{
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
								TypeName: "test_block",
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-block-set": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"test_block": {
							Attributes: map[string]tfsdk.Attribute{
								"test_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						BlockTypes: []*tfprotov6.SchemaNestedBlock{
							{
								Block: &tfprotov6.SchemaBlock{
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
								TypeName: "test_block",
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-block-single": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Blocks: map[string]tfsdk.Block{
						"test_block": {
							Attributes: map[string]tfsdk.Attribute{
								"test_attribute": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSingle,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						BlockTypes: []*tfprotov6.SchemaNestedBlock{
							{
								Block: &tfprotov6.SchemaBlock{
									Attributes: []*tfprotov6.SchemaAttribute{
										{
											Name:     "test_attribute",
											Type:     tftypes.String,
											Required: true,
										},
									},
								},
								Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
								TypeName: "test_block",
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-meta-version": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: &tfsdk.Schema{
					Version: 123,
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ProviderMeta: &tfprotov6.Schema{
					Block:   &tfprotov6.SchemaBlock{},
					Version: 123,
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"resource-multiple-resources": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource_1": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed: true,
								Type:     types.BoolType,
							},
						},
					},
					"test_resource_2": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource_1": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed: true,
									Name:     "test_attribute",
									Type:     tftypes.Bool,
								},
							},
						},
					},
					"test_resource_2": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed: true,
									Name:     "test_attribute",
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed: true,
									Name:     "test_attribute",
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-deprecated": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								DeprecationMessage: "deprecated",
								Optional:           true,
								Type:               types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Deprecated: true,
									Name:       "test_attribute",
									Optional:   true,
									Type:       tftypes.Bool,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-optional": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Optional: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Optional: true,
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-optional-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed: true,
								Optional: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed: true,
									Name:     "test_attribute",
									Optional: true,
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Type:     tftypes.Bool,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-sensitive": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Computed:  true,
								Sensitive: true,
								Type:      types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed:  true,
									Name:      "test_attribute",
									Sensitive: true,
									Type:      tftypes.Bool,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-bool": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.BoolType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.Bool,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-float64": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.Float64Type,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.Number,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-int64": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.Int64Type,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.Number,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-list-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.ListType{
									ElemType: types.ListType{
										ElemType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.List{
										ElementType: tftypes.List{
											ElementType: tftypes.String,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-list-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"test_nested_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name: "test_attribute",
									NestedType: &tfprotov6.SchemaObject{
										Nesting: tfprotov6.SchemaObjectNestingModeList,
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_nested_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-list-object": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.ListType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"test_object_attribute": types.StringType,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.List{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_object_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.List{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-map-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
									"test_nested_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name: "test_attribute",
									NestedType: &tfprotov6.SchemaObject{
										Nesting: tfprotov6.SchemaObjectNestingModeMap,
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_nested_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-map-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.MapType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Map{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-number": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.NumberType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.Number,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-object": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"test_object_attribute": types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_object_attribute": tftypes.String,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-set-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"test_nested_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name: "test_attribute",
									NestedType: &tfprotov6.SchemaObject{
										Nesting: tfprotov6.SchemaObjectNestingModeSet,
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_nested_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-set-object": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.SetType{
									ElemType: types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"test_object_attribute": types.StringType,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Set{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_object_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-set-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.SetType{
									ElemType: types.SetType{
										ElemType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Set{
										ElementType: tftypes.Set{
											ElementType: tftypes.String,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type: types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type: tftypes.Set{
										ElementType: tftypes.String,
									},
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-single-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"test_nested_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name: "test_attribute",
									NestedType: &tfprotov6.SchemaObject{
										Nesting: tfprotov6.SchemaObjectNestingModeSingle,
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_nested_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"resource-attribute-type-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test_attribute": {
								Required: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
			},
		},
		"resource-block-list": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test_block": {
								Attributes: map[string]tfsdk.Attribute{
									"test_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeList,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							BlockTypes: []*tfprotov6.SchemaNestedBlock{
								{
									Block: &tfprotov6.SchemaBlock{
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
									TypeName: "test_block",
								},
							},
						},
					},
				},
			},
		},
		"resource-block-set": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test_block": {
								Attributes: map[string]tfsdk.Attribute{
									"test_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeSet,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							BlockTypes: []*tfprotov6.SchemaNestedBlock{
								{
									Block: &tfprotov6.SchemaBlock{
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
									TypeName: "test_block",
								},
							},
						},
					},
				},
			},
		},
		"resource-block-single": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Blocks: map[string]tfsdk.Block{
							"test_block": {
								Attributes: map[string]tfsdk.Attribute{
									"test_attribute": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: tfsdk.BlockNestingModeSingle,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							BlockTypes: []*tfprotov6.SchemaNestedBlock{
								{
									Block: &tfprotov6.SchemaBlock{
										Attributes: []*tfprotov6.SchemaAttribute{
											{
												Name:     "test_attribute",
												Type:     tftypes.String,
												Required: true,
											},
										},
									},
									Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
									TypeName: "test_block",
								},
							},
						},
					},
				},
			},
		},
		"resource-version": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": &tfsdk.Schema{
						Version: 123,
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block:   &tfprotov6.SchemaBlock{},
						Version: 123,
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.GetProviderSchemaResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
