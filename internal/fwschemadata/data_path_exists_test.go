// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDataPathExists(t *testing.T) {
	t.Parallel()

	type testCase struct {
		data          fwschemadata.Data
		path          path.Path
		expected      bool
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"empty-path": {
			data:     fwschemadata.Data{},
			path:     path.Root("test"),
			expected: false,
		},
		"empty-root": {
			data:     fwschemadata.Data{},
			path:     path.Empty(),
			expected: true,
		},
		"root": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     path.Empty(),
			expected: true,
		},
		"WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     path.Root("test"),
			expected: true,
		},
		"WithAttributeName-mismatch": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			expected: false,
		},
		"WithAttributeName.WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested": tftypes.String,
							},
						},
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"nested": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested": types.StringType,
								},
							},
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("nested"),
			expected: true,
		},
		"WithAttributeName.WithAttributeName-mismatch-child": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested": tftypes.String,
							},
						},
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"nested": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested": types.StringType,
								},
							},
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("other"),
			expected: false,
		},
		"WithAttributeName.WithAttributeName-mismatch-parent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     path.Root("test").AtName("other"),
			expected: false,
		},
		"WithAttributeName.WithElementKeyInt": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: true,
		},
		"WithAttributeName.WithElementKeyInt-mismatch-child": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(1),
			expected: false,
		},
		"WithAttributeName.WithElementKeyInt-mismatch-parent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     path.Root("test").AtListIndex(0),
			expected: false,
		},
		"WithAttributeName.WithElementKeyString": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"key": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("key"),
			expected: true,
		},
		"WithAttributeName.WithElementKeyString-mismatch-child": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"key": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("other"),
			expected: false,
		},
		"WithAttributeName.WithElementKeyString-mismatch-parent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     path.Root("test").AtMapKey("other"),
			expected: false,
		},
		"WithAttributeName.WithElementKeyValue": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.StringValue("testvalue")),
			expected: true,
		},
		"WithAttributeName.WithElementKeyValue-mismatch-child": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.StringValue("othervalue")),
			expected: false,
		},
		"WithAttributeName.WithElementKeyValue-mismatch-parent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     path.Root("test").AtSetValue(types.StringValue("othervalue")),
			expected: false,
		},
		// This is the expected correct path to access a dynamic attribute (at the root only)
		"DynamicType-WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test"),
			expected: true,
		},
		"DynamicType-WithAttributeName-mismatch": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("other"),
			expected: false,
		},
		// This test passes because the underlying `(Data).PathExists` function uses the TerraformValue and not the Schema.
		// Framework dynamic attributes don't allow you to step into them with paths.
		"DynamicType-WithAttributeName.WithAttributeName": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"nested": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("nested"),
			expected: true,
		},
		"DynamicType-WithAttributeName.WithAttributeName-mismatch-child": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"nested": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("other"),
			expected: false,
		},
		"DynamicType-WithAttributeName.WithAttributeName-mismatch-parent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtName("other"),
			expected: false,
		},
		// This test passes because the underlying `(Data).PathExists` function uses the TerraformValue and not the Schema.
		// Framework dynamic attributes don't allow you to step into them with paths.
		"DynamicType-WithAttributeName.WithElementKeyInt": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: true,
		},
		"DynamicType-WithAttributeName.WithElementKeyInt-mismatch-child": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(1),
			expected: false,
		},
		"DynamicType-WithAttributeName.WithElementKeyInt-mismatch-parent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: false,
		},
		// This test passes because the underlying `(Data).PathExists` function uses the TerraformValue and not the Schema.
		// Framework dynamic attributes don't allow you to step into them with paths.
		"DynamicType-WithAttributeName.WithElementKeyString": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"key": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("key"),
			expected: true,
		},
		"DynamicType-WithAttributeName.WithElementKeyString-mismatch-child": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"key": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("other"),
			expected: false,
		},
		"DynamicType-WithAttributeName.WithElementKeyString-mismatch-parent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtMapKey("other"),
			expected: false,
		},
		// This test passes because the underlying `(Data).PathExists` function uses the TerraformValue and not the Schema.
		// Framework dynamic attributes don't allow you to step into them with paths.
		"DynamicType-WithAttributeName.WithElementKeyValue-StringValue": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.StringValue("testvalue")),
			expected: true,
		},
		// This test passes because the underlying `(Data).PathExists` function uses the TerraformValue and not the Schema.
		// Framework dynamic attributes don't allow you to step into them with paths.
		"DynamicType-WithAttributeName.WithElementKeyValue-DynamicValue": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.DynamicValue(types.StringValue("testvalue"))),
			expected: true,
		},
		"DynamicType-WithAttributeName.WithElementKeyValue-mismatch-child": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "testvalue"),
					}),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.StringValue("othervalue")),
			expected: false,
		},
		"DynamicType-WithAttributeName.WithElementKeyValue-mismatch-parent": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.DynamicType,
							Required: true,
						},
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.StringValue("othervalue")),
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := tc.data.PathExists(context.Background(), tc.path)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}
