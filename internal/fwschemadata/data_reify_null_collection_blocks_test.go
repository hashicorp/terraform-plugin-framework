// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDataReifyNullCollectionBlocks(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data          *fwschemadata.Data
		expected      *fwschemadata.Data
		expectedDiags diag.Diagnostics
	}{
		"list-attribute-unmodified": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{}, // intentionally no elements
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"list_attribute": testschema.Attribute{
							Optional: true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_attribute": tftypes.List{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"list_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{}, // intentionally no elements
						),
					},
				),
			},
		},
		"set-attribute-unmodified": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{}, // intentionally no elements
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"set_attribute": testschema.Attribute{
							Optional: true,
							Type: types.SetType{
								ElemType: types.StringType,
							},
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_attribute": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"set_attribute": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{}, // intentionally no elements
						),
					},
				),
			},
		},
		"list-block-null": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"list_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
							nil, // should be converted to empty value
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"list_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{},
						),
					},
				),
			},
		},
		"list-block-unknown": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"list_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"list_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
		},
		"list-block-elements": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"list_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"list_block_attribute": tftypes.String,
										},
									},
									nil, // null value should not matter
								),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"list_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"list_block_attribute": tftypes.String,
										},
									},
									nil, // null value should not matter
								),
							},
						),
					},
				),
			},
		},
		"list-block-zero-elements": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"list_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"list_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"list_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"list_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"list_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{},
						),
					},
				),
			},
		},
		"set-block-null": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"set_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
							nil, // should be converted to empty value
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"set_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{},
						),
					},
				),
			},
		},
		"set-block-unknown": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"set_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"set_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
		},
		"set-block-elements": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"set_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"set_block_attribute": tftypes.String,
										},
									},
									nil, // null value should not matter
								),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"set_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"set_block_attribute": tftypes.String,
										},
									},
									nil, // null value should not matter
								),
							},
						),
					},
				),
			},
		},
		"set-block-zero-elements": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"set_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"set_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"set_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"set_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"set_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"set_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{},
						),
					},
				),
			},
		},
		"single-block-null": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"single_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"single_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_block": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_block": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"single_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"single_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_block": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_block": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
							nil,
						),
					},
				),
			},
		},
		"single-block-unknown": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"single_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"single_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_block": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_block": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"single_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"single_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_block": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_block": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
							tftypes.UnknownValue,
						),
					},
				),
			},
		},
		"single-block-known": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"single_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"single_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_block": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_block": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"single_block_attribute": tftypes.NewValue(tftypes.String, nil),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Blocks: map[string]fwschema.Block{
						"single_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"single_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"single_block": tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
						},
					},
					map[string]tftypes.Value{
						"single_block": tftypes.NewValue(
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"single_block_attribute": tftypes.String,
								},
							},
							map[string]tftypes.Value{
								"single_block_attribute": tftypes.NewValue(tftypes.String, nil),
							},
						),
					},
				),
			},
		},
		// Dynamic attributes that contain underlying list values should be skipped
		"dynamic-attribute-with-list-unmodified": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.DynamicType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "hello"),
							},
						),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"dynamic_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.DynamicType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"dynamic_attribute": tftypes.DynamicPseudoType,
						},
					},
					map[string]tftypes.Value{
						"dynamic_attribute": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.String,
							},
							[]tftypes.Value{
								tftypes.NewValue(tftypes.String, "hello"),
							},
						),
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.data.ReifyNullCollectionBlocks(context.Background())

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.data, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
