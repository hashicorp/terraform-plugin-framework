// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataValueAtPath(t *testing.T) {
	t.Parallel()

	type testCase struct {
		data          fwschemadata.Data
		path          path.Path
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"empty": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, nil),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-nonexistent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "value"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("other"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("other"),
					"Data Read Error",
					"An unexpected error was encountered trying to retrieve type information at a given path. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Error: AttributeName(\"other\") still remains in the path: could not find attribute or block \"other\" in schema",
				),
			},
		},
		"WithAttributeName-List-null-WithElementKeyInt": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: types.StringNull(),
		},
		"WithAttributeName-List-WithElementKeyInt": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "value"),
						tftypes.NewValue(tftypes.String, "othervalue"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-ListNestedAttributes-null-WithElementKeyInt-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Required:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-ListNestedAttributes-null-WithElementKeyInt-WithAttributeName-Object": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"value": tftypes.String,
										},
									},
								},
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"value": tftypes.String,
									},
								},
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.NestedAttribute{
										NestedObject: testschema.NestedAttributeObject{
											Attributes: map[string]fwschema.Attribute{
												"value": testschema.Attribute{
													Type:     types.StringType,
													Optional: true,
												},
											},
										},
										NestingMode: fwschema.NestingModeSingle,
										Optional:    true,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Optional:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.ObjectNull(
				map[string]attr.Type{"value": types.StringType},
			),
		},
		"WithAttributeName-ListNestedAttributes-WithElementKeyInt-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"sub_test": tftypes.NewValue(tftypes.String, "value"),
						}),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeList,
							Required:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-ListNestedBlocks-null-WithElementKeyInt-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.Bool,
								},
							},
						},
						"test": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-ListNestedBlocks-WithElementKeyInt-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.Bool,
								},
							},
						},
						"test": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"sub_test": tftypes.NewValue(tftypes.String, "value"),
						}),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-Map-null-WithElementKeyString": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-Map-WithElementKeyString": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.String, "value"),
						"other":    tftypes.NewValue(tftypes.String, "othervalue"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-Map-WithElementKeyString-nonexistent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.String, "value"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("other"),
			expected: types.StringNull(),
		},
		"WithAttributeName-MapNestedAttributes-null-WithElementKeyInt-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Required:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("element").AtName("sub_test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-MapNestedAttributes-WithElementKeyString-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, map[string]tftypes.Value{
						"element": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"sub_test": tftypes.NewValue(tftypes.String, "value"),
						}),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeMap,
							Required:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("element").AtName("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-Object-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.String, "value"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"sub_test": types.StringType,
								},
							},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-Set-null-WithElementKeyValue": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.StringValue("value")),
			expected: types.StringNull(),
		},
		"WithAttributeName-Set-WithElementKeyValue": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "value"),
						tftypes.NewValue(tftypes.String, "othervalue"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.StringValue("value")),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-SetNestedAttributes-null-WithElementKeyValue-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Required:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: path.Root("test").AtSetValue(types.ObjectValueMust(
				map[string]attr.Type{
					"sub_test": types.StringType,
				},
				map[string]attr.Value{
					"sub_test": types.StringValue("value"),
				},
			)).AtName("sub_test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-SetNestedAttributes-WithElementKeyValue-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"sub_test": tftypes.NewValue(tftypes.String, "value"),
						}),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSet,
							Required:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: path.Root("test").AtSetValue(types.ObjectValueMust(
				map[string]attr.Type{
					"sub_test": types.StringType,
				},
				map[string]attr.Value{
					"sub_test": types.StringValue("value"),
				},
			)).AtName("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-SetNestedBlocks-null-WithElementKeyValue-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.Bool,
								},
							},
						},
						"test": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
			},
			path: path.Root("test").AtSetValue(types.ObjectValueMust(
				map[string]attr.Type{
					"sub_test": types.StringType,
				},
				map[string]attr.Value{
					"sub_test": types.StringValue("value"),
				},
			)).AtName("sub_test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-SetNestedBlocks-WithElementKeyValue-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.Bool,
								},
							},
						},
						"test": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"sub_test": tftypes.String,
								},
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"sub_test": tftypes.NewValue(tftypes.String, "value"),
						}),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
			},
			path: path.Root("test").AtSetValue(types.ObjectValueMust(
				map[string]attr.Type{
					"sub_test": types.StringType,
				},
				map[string]attr.Value{
					"sub_test": types.StringValue("value"),
				},
			)).AtName("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-SingleBlock-null-WithAttributeName-Float32": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Number,
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Bool,
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Number,
						},
					}, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.Float32Type,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Float32Null(),
		},
		"WithAttributeName-SingleBlock-null-WithAttributeName-Float64": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Number,
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Bool,
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Number,
						},
					}, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.Float64Type,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Float64Null(),
		},
		"WithAttributeName-SingleBlock-null-WithAttributeName-Int32": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Number,
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Bool,
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Number,
						},
					}, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.Int32Type,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Int32Null(),
		},
		"WithAttributeName-SingleBlock-null-WithAttributeName-Int64": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Number,
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Bool,
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Number,
						},
					}, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.Int64Type,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Int64Null(),
		},
		"WithAttributeName-SingleBlock-null-WithAttributeName-Set": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Set{
									ElementType: tftypes.Bool,
								},
							},
						},
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Set{
									ElementType: tftypes.String,
								},
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Set{
								ElementType: tftypes.Bool,
							},
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					}, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type: types.SetType{
											ElemType: types.BoolType,
										},
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type: types.SetType{
											ElemType: types.StringType,
										},
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.SetNull(types.StringType),
		},
		"WithAttributeName-SingleBlock-null-WithAttributeName-String": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Bool,
						},
					}, nil),
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.String,
						},
					}, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-SingleBlock-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"other_attr": tftypes.Bool,
						"other_block": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Bool,
							},
						},
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
					},
				}, map[string]tftypes.Value{
					"other_attr": tftypes.NewValue(tftypes.Bool, nil),
					"other_block": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.Bool, true),
					}),
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.String, "value"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"other_attr": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"other_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.BoolType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
						"test": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-SingleNestedAttributes-null-WithAttributeName-Float32": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Number,
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Number,
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.Float32Type,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Float32Null(),
		},
		"WithAttributeName-SingleNestedAttributes-null-WithAttributeName-Float64": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Number,
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Number,
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.Float64Type,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Float64Null(),
		},
		"WithAttributeName-SingleNestedAttributes-null-WithAttributeName-Int32": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Number,
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Number,
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.Int32Type,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Int32Null(),
		},
		"WithAttributeName-SingleNestedAttributes-null-WithAttributeName-Int64": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Number,
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Number,
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.Int64Type,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Int64Null(),
		},
		"WithAttributeName-SingleNestedAttributes-null-WithAttributeName-Set": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.Set{
									ElementType: tftypes.String,
								},
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.Set{
								ElementType: tftypes.String,
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type: types.SetType{
											ElemType: types.StringType,
										},
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Optional:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.SetNull(types.StringType),
		},
		"WithAttributeName-SingleNestedAttributes-null-WithAttributeName-String": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.String,
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Required:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-SingleNestedAttributes-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.String, "value"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.NestedAttribute{
							NestedObject: testschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_test": testschema.Attribute{
										Type:     types.StringType,
										Required: true,
									},
								},
							},
							NestingMode: fwschema.NestingModeSingle,
							Required:    true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-String-null": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.StringNull(),
		},
		"WithAttributeName-String-unknown": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.StringUnknown(),
		},
		"WithAttributeName-String-value": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.StringValue("value"),
		},
		"WithAttributeName-Dynamic-null": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.DynamicPseudoType,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.DynamicPseudoType, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.DynamicNull(),
		},
		"WithAttributeName-Dynamic-underlying-value-null": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.DynamicPseudoType,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, nil), // A concrete type! :O
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.DynamicValue(types.StringNull()),
		},
		"WithAttributeName-Dynamic-unknown": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.DynamicPseudoType,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.DynamicUnknown(),
		},
		"WithAttributeName-Dynamic-underlying-value-unknown": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.DynamicPseudoType,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), // A concrete type! :O
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.DynamicValue(types.NumberUnknown()),
		},
		"WithAttributeName-Dynamic-value": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.DynamicValue(types.StringValue("value")),
		},
		// MAINTAINER NOTE: Paths currently cannot target values inside of dynamic types, even if the underlying data matches the path.
		// If we enable this functionality in the future, this test should be updated to correctly grab the data.
		"WithAttributeName-Dynamic-List-WithElementKeyInt-Error": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "value"),
						tftypes.NewValue(tftypes.String, "othervalue"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Data Read Error",
					"An unexpected error was encountered trying to retrieve type information at a given path. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Error: path leads to element or attribute nested in a schema.DynamicAttribute",
				),
			},
		},
		// MAINTAINER NOTE: Paths currently cannot target values inside of dynamic types, even if the underlying data matches the path.
		// If we enable this functionality in the future, this test should be updated to correctly grab the data.
		"WithAttributeName-Dynamic-Map-WithElementKeyString-Error": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.String, "value"),
						"other":    tftypes.NewValue(tftypes.String, "othervalue"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("sub_test"),
					"Data Read Error",
					"An unexpected error was encountered trying to retrieve type information at a given path. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Error: path leads to element or attribute nested in a schema.DynamicAttribute",
				),
			},
		},
		// MAINTAINER NOTE: Paths currently cannot target values inside of dynamic types, even if the underlying data matches the path.
		// If we enable this functionality in the future, this test should be updated to correctly grab the data.
		"WithAttributeName-Dynamic-Set-WithElementKeyValue-At-DynamicValue-Error": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.DynamicPseudoType,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "value"),
						tftypes.NewValue(tftypes.String, "othervalue"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.DynamicValue(types.StringValue("value"))),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.DynamicValue(types.StringValue("value"))),
					"Data Read Error",
					"An unexpected error was encountered trying to retrieve type information at a given path. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Error: path leads to element or attribute nested in a schema.DynamicAttribute",
				),
			},
		},
		// MAINTAINER NOTE: Paths currently cannot target values inside of dynamic types, even if the underlying data matches the path.
		// If we enable this functionality in the future, this test should be updated to correctly grab the data.
		"WithAttributeName-Dynamic-Object-WithAttributeName-Error": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"sub_test": tftypes.String,
							},
						},
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.String, "value"),
					}),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("sub_test"),
					"Data Read Error",
					"An unexpected error was encountered trying to retrieve type information at a given path. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Error: path leads to element or attribute nested in a schema.DynamicAttribute",
				),
			},
		},
		"AttrTypeWithValidateError": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:          path.Root("test"),
			expected:      nil,
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(path.Root("test"))},
		},
		"AttrTypeWithValidateAttributeError": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateAttributeError{},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:          path.Root("test"),
			expected:      nil,
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(path.Root("test"))},
		},
		"AttrTypeWithValidateWarning": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:          path.Root("test"),
			expected:      testtypes.String{InternalString: types.StringValue("value"), CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(path.Root("test"))},
		},
		"AttrTypeWithValidateAttributeWarning": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateAttributeWarning{},
							Required: true,
						},
						"other": testschema.Attribute{
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:          path.Root("test"),
			expected:      testtypes.StringValueWithValidateAttributeWarning{InternalString: testtypes.String{InternalString: types.StringValue("value"), CreatedBy: testtypes.StringTypeWithValidateAttributeWarning{}}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(path.Root("test"))},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			val, diags := tc.data.ValueAtPath(context.Background(), tc.path)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
