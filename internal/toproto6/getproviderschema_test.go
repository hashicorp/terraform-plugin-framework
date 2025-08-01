// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
		"action-multiple-actions": {
			input: &fwserver.GetProviderSchemaResponse{
				ActionSchemas: map[string]actionschema.SchemaType{
					"test_action_1": actionschema.UnlinkedSchema{
						Attributes: map[string]actionschema.Attribute{
							"test_attribute": actionschema.StringAttribute{
								Required: true,
							},
						},
					},
					"test_action_2": actionschema.UnlinkedSchema{
						Attributes: map[string]actionschema.Attribute{
							"test_attribute": actionschema.StringAttribute{
								Optional:           true,
								DeprecationMessage: "deprecated",
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{
					"test_action_1": {
						Type: tfprotov6.UnlinkedActionSchemaType{},
						Schema: &tfprotov6.Schema{
							Version: 0,
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "test_attribute",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
						},
					},
					"test_action_2": {
						Type: tfprotov6.UnlinkedActionSchemaType{},
						Schema: &tfprotov6.Schema{
							Version: 0,
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:       "test_attribute",
										Type:       tftypes.String,
										Optional:   true,
										Deprecated: true,
									},
								},
							},
						},
					},
				},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"action-type-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				ActionSchemas: map[string]actionschema.SchemaType{
					"test_action": actionschema.UnlinkedSchema{
						Attributes: map[string]actionschema.Attribute{
							"test_attribute": actionschema.SingleNestedAttribute{
								Attributes: map[string]actionschema.Attribute{
									"test_nested_attribute": actionschema.StringAttribute{
										Required: true,
									},
								},
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{
					"test_action": {
						Type: tfprotov6.UnlinkedActionSchemaType{},
						Schema: &tfprotov6.Schema{
							Version: 0,
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name: "test_attribute",
										NestedType: &tfprotov6.SchemaObject{
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:     "test_nested_attribute",
													Type:     tftypes.String,
													Required: true,
												},
											},
											Nesting: tfprotov6.SchemaObjectNestingModeSingle,
										},
										Required: true,
									},
								},
							},
						},
					},
				},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-multiple-data-sources": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source_1": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Computed: true,
							},
						},
					},
					"test_data_source_2": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-deprecated": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								DeprecationMessage: "deprecated",
								Optional:           true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-optional": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-optional-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Computed: true,
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-sensitive": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Computed:  true,
								Sensitive: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-write-only": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Computed:  true,
									Name:      "test_attribute",
									WriteOnly: false,
									Type:      tftypes.Bool,
								},
							},
						},
					},
				},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-bool": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-float64": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.Float64Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-int32": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.Int32Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-int64": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.Int64Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-list-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.ListAttribute{
								Required: true,
								ElementType: types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-list-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.ListNestedAttribute{
								NestedObject: datasourceschema.NestedAttributeObject{
									Attributes: map[string]datasourceschema.Attribute{
										"test_nested_attribute": datasourceschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-list-object": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.ListAttribute{
								Required: true,
								ElementType: types.ObjectType{
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
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.ListAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-map-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.MapNestedAttribute{
								NestedObject: datasourceschema.NestedAttributeObject{
									Attributes: map[string]datasourceschema.Attribute{
										"test_nested_attribute": datasourceschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-map-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.MapAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-number": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.NumberAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-object": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.ObjectAttribute{
								Required: true,
								AttributeTypes: map[string]attr.Type{
									"test_object_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-set-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.SetNestedAttribute{
								NestedObject: datasourceschema.NestedAttributeObject{
									Attributes: map[string]datasourceschema.Attribute{
										"test_nested_attribute": datasourceschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-set-object": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.SetAttribute{
								Required: true,
								ElementType: types.ObjectType{
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
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-set-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.SetAttribute{
								Required: true,
								ElementType: types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-single-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.SingleNestedAttribute{
								Attributes: map[string]datasourceschema.Attribute{
									"test_nested_attribute": datasourceschema.StringAttribute{
										Required: true,
									},
								},
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-string": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-attribute-type-dynamic": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Attributes: map[string]datasourceschema.Attribute{
							"test_attribute": datasourceschema.DynamicAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{
					"test_data_source": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.DynamicPseudoType,
								},
							},
						},
					},
				},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-block-list": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Blocks: map[string]datasourceschema.Block{
							"test_block": datasourceschema.ListNestedBlock{
								NestedObject: datasourceschema.NestedBlockObject{
									Attributes: map[string]datasourceschema.Attribute{
										"test_attribute": datasourceschema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-block-set": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Blocks: map[string]datasourceschema.Block{
							"test_block": datasourceschema.SetNestedBlock{
								NestedObject: datasourceschema.NestedBlockObject{
									Attributes: map[string]datasourceschema.Attribute{
										"test_attribute": datasourceschema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"data-source-block-single": {
			input: &fwserver.GetProviderSchemaResponse{
				DataSourceSchemas: map[string]fwschema.Schema{
					"test_data_source": datasourceschema.Schema{
						Blocks: map[string]datasourceschema.Block{
							"test_block": datasourceschema.SingleNestedBlock{
								Attributes: map[string]datasourceschema.Attribute{
									"test_attribute": datasourceschema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas: map[string]*tfprotov6.ActionSchema{},
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
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas:          map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-multiple-ephemeral-resources": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource_1": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								Computed: true,
							},
						},
					},
					"test_ephemeral_resource_2": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource_1": {
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
					"test_ephemeral_resource_2": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-deprecated": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								DeprecationMessage: "deprecated",
								Optional:           true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-optional": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-optional-computed": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								Computed: true,
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-sensitive": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								Computed:  true,
								Sensitive: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-bool": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-float32": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.Float32Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-float64": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.Float64Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-int32": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.Int32Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-int64": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.Int64Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-list-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.ListAttribute{
								Required: true,
								ElementType: types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-list-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.ListNestedAttribute{
								NestedObject: ephemeralschema.NestedAttributeObject{
									Attributes: map[string]ephemeralschema.Attribute{
										"test_nested_attribute": ephemeralschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-list-object": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.ListAttribute{
								Required: true,
								ElementType: types.ObjectType{
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
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-list-string": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.ListAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-map-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.MapNestedAttribute{
								NestedObject: ephemeralschema.NestedAttributeObject{
									Attributes: map[string]ephemeralschema.Attribute{
										"test_nested_attribute": ephemeralschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-map-string": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.MapAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-number": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.NumberAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-object": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.ObjectAttribute{
								Required: true,
								AttributeTypes: map[string]attr.Type{
									"test_object_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-set-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.SetNestedAttribute{
								NestedObject: ephemeralschema.NestedAttributeObject{
									Attributes: map[string]ephemeralschema.Attribute{
										"test_nested_attribute": ephemeralschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-set-object": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.SetAttribute{
								Required: true,
								ElementType: types.ObjectType{
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
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-set-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.SetAttribute{
								Required: true,
								ElementType: types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-set-string": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-single-nested-attributes": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.SingleNestedAttribute{
								Attributes: map[string]ephemeralschema.Attribute{
									"test_nested_attribute": ephemeralschema.StringAttribute{
										Required: true,
									},
								},
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-string": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-attribute-type-dynamic": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Attributes: map[string]ephemeralschema.Attribute{
							"test_attribute": ephemeralschema.DynamicAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.DynamicPseudoType,
								},
							},
						},
					},
				},
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-block-list": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Blocks: map[string]ephemeralschema.Block{
							"test_block": ephemeralschema.ListNestedBlock{
								NestedObject: ephemeralschema.NestedBlockObject{
									Attributes: map[string]ephemeralschema.Attribute{
										"test_attribute": ephemeralschema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-block-set": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Blocks: map[string]ephemeralschema.Block{
							"test_block": ephemeralschema.SetNestedBlock{
								NestedObject: ephemeralschema.NestedBlockObject{
									Attributes: map[string]ephemeralschema.Attribute{
										"test_attribute": ephemeralschema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"ephemeral-resource-block-single": {
			input: &fwserver.GetProviderSchemaResponse{
				EphemeralResourceSchemas: map[string]fwschema.Schema{
					"test_ephemeral_resource": ephemeralschema.Schema{
						Blocks: map[string]ephemeralschema.Block{
							"test_block": ephemeralschema.SingleNestedBlock{
								Attributes: map[string]ephemeralschema.Attribute{
									"test_attribute": ephemeralschema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:     map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas: map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{
					"test_ephemeral_resource": {
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
				Functions:           map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"functions": {
			input: &fwserver.GetProviderSchemaResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction1": {
						Return: function.StringReturn{},
					},
					"testfunction2": {
						Return: function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions: map[string]*tfprotov6.Function{
					"testfunction1": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
					"testfunction2": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"functions-deprecationmessage": {
			input: &fwserver.GetProviderSchemaResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						DeprecationMessage: "test deprecation message",
						Return:             function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						DeprecationMessage: "test deprecation message",
						Parameters:         []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"functions-description": {
			input: &fwserver.GetProviderSchemaResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Description: "test description",
						Return:      function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Description: "test description",
						Parameters:  []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"functions-parameters": {
			input: &fwserver.GetProviderSchemaResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Parameters: []function.Parameter{
							function.BoolParameter{},
							function.Int64Parameter{},
							function.StringParameter{},
						},
						Return: function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Parameters: []*tfprotov6.FunctionParameter{
							{
								Type: tftypes.Bool,
							},
							{
								Type: tftypes.Number,
							},
							{
								Type: tftypes.String,
							},
						},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"functions-result": {
			input: &fwserver.GetProviderSchemaResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Return: function.StringReturn{},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
					},
				},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"functions-summary": {
			input: &fwserver.GetProviderSchemaResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Return:  function.StringReturn{},
						Summary: "test summary",
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
						Summary: "test summary",
					},
				},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"functions-variadicparameter": {
			input: &fwserver.GetProviderSchemaResponse{
				FunctionDefinitions: map[string]function.Definition{
					"testfunction": {
						Return:            function.StringReturn{},
						VariadicParameter: function.StringParameter{},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions: map[string]*tfprotov6.Function{
					"testfunction": {
						Parameters: []*tfprotov6.FunctionParameter{},
						Return: &tfprotov6.FunctionReturn{
							Type: tftypes.String,
						},
						VariadicParameter: &tfprotov6.FunctionParameter{
							Type: tftypes.String,
						},
					},
				},
				ListResourceSchemas: map[string]*tfprotov6.Schema{},
				ResourceSchemas:     map[string]*tfprotov6.Schema{},
			},
		},
		"list-resource-multiple-list-resources": {
			input: &fwserver.GetProviderSchemaResponse{
				ListResourceSchemas: map[string]fwschema.Schema{
					"test_list_resource_1": listschema.Schema{
						Attributes: map[string]listschema.Attribute{
							"test_attribute": listschema.StringAttribute{
								Optional: true,
							},
						},
					},
					"test_list_resource_2": listschema.Schema{
						Attributes: map[string]listschema.Attribute{
							"test_attribute": ephemeralschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{
					"test_list_resource_1": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Optional: true,
									Name:     "test_attribute",
									Type:     tftypes.String,
								},
							},
						},
					},
					"test_list_resource_2": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Optional: true,
									Name:     "test_attribute",
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"list-resource-attribute-deprecated": {
			input: &fwserver.GetProviderSchemaResponse{
				ListResourceSchemas: map[string]fwschema.Schema{
					"test_list_resource": listschema.Schema{
						Attributes: map[string]listschema.Attribute{
							"test_attribute": listschema.StringAttribute{
								DeprecationMessage: "deprecated",
								Optional:           true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{
					"test_list_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Deprecated: true,
									Name:       "test_attribute",
									Optional:   true,
									Type:       tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"list-resource-attribute-optional": {
			input: &fwserver.GetProviderSchemaResponse{
				ListResourceSchemas: map[string]fwschema.Schema{
					"test_list_resource": listschema.Schema{
						Attributes: map[string]listschema.Attribute{
							"test_attribute": listschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{
					"test_list_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Optional: true,
									Type:     tftypes.String,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"list-resource-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				ListResourceSchemas: map[string]fwschema.Schema{
					"test_list_resource": listschema.Schema{
						Attributes: map[string]listschema.Attribute{
							"test_attribute": listschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{
					"test_list_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Type:     tftypes.String,
									Required: true,
								},
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"list-resource-attribute-type-string": {
			input: &fwserver.GetProviderSchemaResponse{
				ListResourceSchemas: map[string]fwschema.Schema{
					"test_list_resource": listschema.Schema{
						Attributes: map[string]listschema.Attribute{
							"test_attribute": listschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas: map[string]*tfprotov6.Schema{
					"test_list_resource": {
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
		"provider-attribute-deprecated": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.BoolAttribute{
							DeprecationMessage: "deprecated",
							Optional:           true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.BoolAttribute{
							Optional: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"provider-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.BoolAttribute{
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:      "test_attribute",
								Optional:  true,
								Sensitive: true,
								Type:      tftypes.Bool,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-attribute-write-only": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.BoolAttribute{
							Optional: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:      "test_attribute",
								Optional:  true,
								WriteOnly: false,
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.Float64Attribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"provider-attribute-type-int32": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.Int32Attribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.ListAttribute{
							Required: true,
							ElementType: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.ListNestedAttribute{
							NestedObject: providerschema.NestedAttributeObject{
								Attributes: map[string]providerschema.Attribute{
									"test_nested_attribute": providerschema.StringAttribute{
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.ListAttribute{
							Required: true,
							ElementType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"test_object_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.MapNestedAttribute{
							NestedObject: providerschema.NestedAttributeObject{
								Attributes: map[string]providerschema.Attribute{
									"test_nested_attribute": providerschema.StringAttribute{
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.MapAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.NumberAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.ObjectAttribute{
							Required: true,
							AttributeTypes: map[string]attr.Type{
								"test_object_attribute": types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.SetNestedAttribute{
							NestedObject: providerschema.NestedAttributeObject{
								Attributes: map[string]providerschema.Attribute{
									"test_nested_attribute": providerschema.StringAttribute{
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.SetAttribute{
							Required: true,
							ElementType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"test_object_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.SetAttribute{
							Required: true,
							ElementType: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.SingleNestedAttribute{
							Attributes: map[string]providerschema.Attribute{
								"test_nested_attribute": providerschema.StringAttribute{
									Required: true,
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"provider-attribute-type-dynamic": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: providerschema.Schema{
					Attributes: map[string]providerschema.Attribute{
						"test_attribute": providerschema.DynamicAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				Provider: &tfprotov6.Schema{
					Block: &tfprotov6.SchemaBlock{
						Attributes: []*tfprotov6.SchemaAttribute{
							{
								Name:     "test_attribute",
								Required: true,
								Type:     tftypes.DynamicPseudoType,
							},
						},
					},
				},
				ResourceSchemas: map[string]*tfprotov6.Schema{},
			},
		},
		"provider-block-list": {
			input: &fwserver.GetProviderSchemaResponse{
				Provider: providerschema.Schema{
					Blocks: map[string]providerschema.Block{
						"test_block": providerschema.ListNestedBlock{
							NestedObject: providerschema.NestedBlockObject{
								Attributes: map[string]providerschema.Attribute{
									"test_attribute": providerschema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Blocks: map[string]providerschema.Block{
						"test_block": providerschema.SetNestedBlock{
							NestedObject: providerschema.NestedBlockObject{
								Attributes: map[string]providerschema.Attribute{
									"test_attribute": providerschema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				Provider: providerschema.Schema{
					Blocks: map[string]providerschema.Block{
						"test_block": providerschema.SingleNestedBlock{
							Attributes: map[string]providerschema.Attribute{
								"test_attribute": providerschema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"provider-meta-attribute-optional": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.BoolAttribute{
							Optional: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"provider-meta-attribute-required": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"provider-meta-attribute-type-bool": {
			input: &fwserver.GetProviderSchemaResponse{
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.Float64Attribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.ListAttribute{
							Required: true,
							ElementType: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.ListNestedAttribute{
							NestedObject: metaschema.NestedAttributeObject{
								Attributes: map[string]metaschema.Attribute{
									"test_nested_attribute": metaschema.StringAttribute{
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.ListAttribute{
							Required: true,
							ElementType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"test_nested_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.MapNestedAttribute{
							NestedObject: metaschema.NestedAttributeObject{
								Attributes: map[string]metaschema.Attribute{
									"test_nested_attribute": metaschema.StringAttribute{
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.MapAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.NumberAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.ObjectAttribute{
							Required: true,
							AttributeTypes: map[string]attr.Type{
								"test_object_attribute": types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.SetNestedAttribute{
							NestedObject: metaschema.NestedAttributeObject{
								Attributes: map[string]metaschema.Attribute{
									"test_nested_attribute": metaschema.StringAttribute{
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.SetAttribute{
							Required: true,
							ElementType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"test_object_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.SetAttribute{
							Required: true,
							ElementType: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.SingleNestedAttribute{
							Attributes: map[string]metaschema.Attribute{
								"test_nested_attribute": metaschema.StringAttribute{
									Required: true,
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
				ProviderMeta: metaschema.Schema{
					Attributes: map[string]metaschema.Attribute{
						"test_attribute": metaschema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"resource-multiple-resources": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource_1": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Computed: true,
							},
						},
					},
					"test_resource_2": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								DeprecationMessage: "deprecated",
								Optional:           true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Computed: true,
								Optional: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Computed:  true,
								Sensitive: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"resource-attribute-write-only": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Optional:  true,
								WriteOnly: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Optional:  true,
									Name:      "test_attribute",
									WriteOnly: true,
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.Float64Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"resource-attribute-type-int32": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.Int32Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.Int64Attribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.ListAttribute{
								Required: true,
								ElementType: types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.ListNestedAttribute{
								NestedObject: resourceschema.NestedAttributeObject{
									Attributes: map[string]resourceschema.Attribute{
										"test_nested_attribute": resourceschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.ListAttribute{
								Required: true,
								ElementType: types.ObjectType{
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
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.ListAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.MapNestedAttribute{
								NestedObject: resourceschema.NestedAttributeObject{
									Attributes: map[string]resourceschema.Attribute{
										"test_nested_attribute": resourceschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.MapAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.NumberAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.ObjectAttribute{
								Required: true,
								AttributeTypes: map[string]attr.Type{
									"test_object_attribute": types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.SetNestedAttribute{
								NestedObject: resourceschema.NestedAttributeObject{
									Attributes: map[string]resourceschema.Attribute{
										"test_nested_attribute": resourceschema.StringAttribute{
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
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.SetAttribute{
								Required: true,
								ElementType: types.ObjectType{
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
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.SetAttribute{
								Required: true,
								ElementType: types.SetType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.SetAttribute{
								Required:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.SingleNestedAttribute{
								Attributes: map[string]resourceschema.Attribute{
									"test_nested_attribute": resourceschema.StringAttribute{
										Required: true,
									},
								},
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		"resource-attribute-type-dynamic": {
			input: &fwserver.GetProviderSchemaResponse{
				ResourceSchemas: map[string]fwschema.Schema{
					"test_resource": resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"test_attribute": resourceschema.DynamicAttribute{
								Required: true,
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
				ResourceSchemas: map[string]*tfprotov6.Schema{
					"test_resource": {
						Block: &tfprotov6.SchemaBlock{
							Attributes: []*tfprotov6.SchemaAttribute{
								{
									Name:     "test_attribute",
									Required: true,
									Type:     tftypes.DynamicPseudoType,
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
					"test_resource": resourceschema.Schema{
						Blocks: map[string]resourceschema.Block{
							"test_block": resourceschema.ListNestedBlock{
								NestedObject: resourceschema.NestedBlockObject{
									Attributes: map[string]resourceschema.Attribute{
										"test_attribute": resourceschema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Blocks: map[string]resourceschema.Block{
							"test_block": resourceschema.SetNestedBlock{
								NestedObject: resourceschema.NestedBlockObject{
									Attributes: map[string]resourceschema.Attribute{
										"test_attribute": resourceschema.StringAttribute{
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Blocks: map[string]resourceschema.Block{
							"test_block": resourceschema.SingleNestedBlock{
								Attributes: map[string]resourceschema.Attribute{
									"test_attribute": resourceschema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
					"test_resource": resourceschema.Schema{
						Version: 123,
					},
				},
			},
			expected: &tfprotov6.GetProviderSchemaResponse{
				ActionSchemas:            map[string]*tfprotov6.ActionSchema{},
				DataSourceSchemas:        map[string]*tfprotov6.Schema{},
				EphemeralResourceSchemas: map[string]*tfprotov6.Schema{},
				Functions:                map[string]*tfprotov6.Function{},
				ListResourceSchemas:      map[string]*tfprotov6.Schema{},
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := toproto6.GetProviderSchemaResponse(context.Background(), testCase.input)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
