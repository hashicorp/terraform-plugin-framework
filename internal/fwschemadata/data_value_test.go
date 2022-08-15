package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.StringType,
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: nil,
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
										"value": {
											Type:     types.StringType,
											Optional: true,
										},
									}),
									Optional: true,
								},
							}),
							Optional: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.Object{
				Null:      true,
				AttrTypes: map[string]attr.Type{"value": types.StringType},
			},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"other_attr": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]tfsdk.Block{
						"other_block": {
							Attributes: map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.BoolType,
									Optional: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
						},
						"test": {
							Attributes: map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"other_attr": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]tfsdk.Block{
						"other_block": {
							Attributes: map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.BoolType,
									Optional: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
						},
						"test": {
							Attributes: map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeList,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0).AtName("sub_test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("other"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("element").AtName("sub_test"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("element").AtName("sub_test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"sub_test": types.StringType,
								},
							},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.String{Value: "value"}),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.String{Value: "value"}),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: path.Root("test").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"sub_test": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"sub_test": types.String{Value: "value"},
				},
			}).AtName("sub_test"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: path.Root("test").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"sub_test": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"sub_test": types.String{Value: "value"},
				},
			}).AtName("sub_test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"other_attr": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]tfsdk.Block{
						"other_block": {
							Attributes: map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.BoolType,
									Optional: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
						},
						"test": {
							Attributes: map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
						},
					},
				},
			},
			path: path.Root("test").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"sub_test": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"sub_test": types.String{Value: "value"},
				},
			}).AtName("sub_test"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"other_attr": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
					Blocks: map[string]tfsdk.Block{
						"other_block": {
							Attributes: map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.BoolType,
									Optional: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
						},
						"test": {
							Attributes: map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: tfsdk.BlockNestingModeSet,
						},
					},
				},
			},
			path: path.Root("test").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"sub_test": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"sub_test": types.String{Value: "value"},
				},
			}).AtName("sub_test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.Float64Type,
									Optional: true,
								},
							}),
							Optional: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Float64{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.Int64Type,
									Optional: true,
								},
							}),
							Optional: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Int64{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type: types.SetType{
										ElemType: types.StringType,
									},
									Optional: true,
								},
							}),
							Optional: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.Set{ElemType: types.StringType, Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("sub_test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.StringType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.String{Null: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.StringType,
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.String{Unknown: true},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.StringType,
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: types.String{Value: "value"},
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
						"other": {
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:          path.Root("test"),
			expected:      testtypes.String{InternalString: types.String{Value: "value"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(path.Root("test"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
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
