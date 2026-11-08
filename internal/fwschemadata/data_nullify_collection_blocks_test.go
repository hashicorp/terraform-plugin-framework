// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDataNullifyCollectionBlocks(t *testing.T) {
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
							nil,
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
							nil,
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
							[]tftypes.Value{}, // intentionally no elements
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
							nil, // should be converted to null value
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
							nil,
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
							nil,
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
							[]tftypes.Value{}, // intentionally no elements
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
							nil, // should be converted to null value
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.data.NullifyCollectionBlocks(context.Background())

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.data, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

// BenchmarkDataCollectionBlocks uses a large recursive schema with a small value,
// like a rule selecting one statement from many supported statement types.
func BenchmarkDataCollectionBlocks(b *testing.B) {
	ctx := b.Context()
	for _, depth := range []int{0, 3, 6} {
		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			var nested func(int) testschema.NestedBlockObject
			nested = func(depth int) testschema.NestedBlockObject {
				n := testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"value": testschema.Attribute{Type: types.StringType, Optional: true},
					},
				}
				if depth > 0 {
					n.Blocks = map[string]fwschema.Block{}
					child := nested(depth - 1)
					for _, name := range []string{"and", "or", "not"} {
						n.Blocks[name] = testschema.Block{
							NestingMode: fwschema.BlockNestingModeList, NestedObject: child,
							CustomType: types.ListType{ElemType: child.Type()},
						}
					}
				}
				return n
			}
			root := nested(depth)
			s := testschema.Schema{Blocks: map[string]fwschema.Block{
				"statement": testschema.Block{
					NestingMode:  fwschema.BlockNestingModeList,
					NestedObject: root,
					CustomType:   types.ListType{ElemType: root.Type()},
				},
			}}
			ty := s.Type().TerraformType(ctx)
			json := `{"statement":[` + strings.Repeat(`{"and":[`, depth) + `{"value":"test"}` + strings.Repeat(`]}`, depth) + `]}`
			v, err := tftypes.ValueFromJSONWithOpts([]byte(json), ty, tftypes.ValueFromJSONOpts{})
			if err != nil {
				b.Fatal(err)
			}
			for _, reify := range []bool{false, true} {
				name := "nullify"
				input := v
				if reify {
					name = "reify"
				} else {
					d := fwschemadata.Data{Schema: s, TerraformValue: v}
					if diags := d.ReifyNullCollectionBlocks(ctx); diags.HasError() {
						b.Fatal(diags)
					}
					input = d.TerraformValue
				}
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						d := fwschemadata.Data{Schema: s, TerraformValue: input}
						var diags diag.Diagnostics
						if reify {
							diags = d.ReifyNullCollectionBlocks(ctx)
						} else {
							diags = d.NullifyCollectionBlocks(ctx)
						}
						if diags.HasError() {
							b.Fatal(diags)
						}
					}
				})
			}
		})
	}
}
