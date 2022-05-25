package tfsdk

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSchemaAttributeAtPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema      Schema
		path        *tftypes.AttributePath
		expected    Attribute
		expectedErr string
	}{
		"empty-root": {
			schema:      Schema{},
			path:        tftypes.NewAttributePath(),
			expected:    Attribute{},
			expectedErr: "got unexpected type tfsdk.Schema",
		},
		"empty-nil": {
			schema:      Schema{},
			path:        nil,
			expected:    Attribute{},
			expectedErr: "got unexpected type tfsdk.Schema",
		},
		"root": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath(),
			expected:    Attribute{},
			expectedErr: "got unexpected type tfsdk.Schema",
		},
		"nil": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path:        nil,
			expected:    Attribute{},
			expectedErr: "got unexpected type tfsdk.Schema",
		},
		"WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-ListNestedAttributes-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: ListNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected:    Attribute{},
			expectedErr: "AttributeName(\"sub_test\") still remains in the path: can't apply tftypes.AttributeName to ListNestedAttributes",
		},
		"WithAttributeName-ListNestedAttributes-WithElementKeyInt": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: ListNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithAttributeName-ListNestedAttributes-WithElementKeyInt-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: ListNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-ListNestedAttributes-WithElementKeyString": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: ListNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected:    Attribute{},
			expectedErr: "ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to ListNestedAttributes",
		},
		"WithAttributeName-ListNestedAttributes-WithElementKeyValue": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: ListNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "sub_test")),
			expected:    Attribute{},
			expectedErr: "ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't apply tftypes.ElementKeyValue to ListNestedAttributes",
		},
		"WithAttributeName-ListNestedBlocks-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected:    Attribute{},
			expectedErr: "AttributeName(\"sub_test\") still remains in the path: can't apply tftypes.AttributeName to block NestingModeList",
		},
		"WithAttributeName-ListNestedBlocks-WithElementKeyInt": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithAttributeName-ListNestedBlocks-WithElementKeyInt-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-ListNestedBlocks-WithElementKeyString": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected:    Attribute{},
			expectedErr: "ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to block NestingModeList",
		},
		"WithAttributeName-ListNestedBlocks-WithElementKeyValue": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "sub_test")),
			expected:    Attribute{},
			expectedErr: "ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't apply tftypes.ElementKeyValue to block NestingModeList",
		},
		"WithAttributeName-MapNestedAttributes-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: MapNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected:    Attribute{},
			expectedErr: "AttributeName(\"sub_test\") still remains in the path: can't use tftypes.AttributeName on maps",
		},
		"WithAttributeName-MapNestedAttributes-WithElementKeyInt": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: MapNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: "ElementKeyInt(0) still remains in the path: can't use tftypes.ElementKeyInt on maps",
		},
		"WithAttributeName-MapNestedAttributes-WithElementKeyString": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: MapNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithAttributeName-MapNestedAttributes-WithElementKeyString-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: MapNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("element").WithAttributeName("sub_test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-MapNestedAttributes-WithElementKeyValue": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: MapNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "sub_test")),
			expected:    Attribute{},
			expectedErr: "ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't use tftypes.ElementKeyValue on maps",
		},
		"WithAttributeName-SetNestedAttributes-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SetNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected:    Attribute{},
			expectedErr: "AttributeName(\"sub_test\") still remains in the path: can't use tftypes.AttributeName on sets",
		},
		"WithAttributeName-SetNestedAttributes-WithElementKeyInt": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SetNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: "ElementKeyInt(0) still remains in the path: can't use tftypes.ElementKeyInt on sets",
		},
		"WithAttributeName-SetNestedAttributes-WithElementKeyString": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SetNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected:    Attribute{},
			expectedErr: "ElementKeyString(\"sub_test\") still remains in the path: can't use tftypes.ElementKeyString on sets",
		},
		"WithAttributeName-SetNestedAttributes-WithElementKeyValue": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SetNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "sub_test")),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithAttributeName-SetNestedAttributes-WithElementKeyValue-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SetNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "element")).WithAttributeName("sub_test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-SetNestedBlocks-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected:    Attribute{},
			expectedErr: "AttributeName(\"sub_test\") still remains in the path: can't apply tftypes.AttributeName to block NestingModeSet",
		},
		"WithAttributeName-SetNestedBlocks-WithElementKeyInt": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: "ElementKeyInt(0) still remains in the path: can't apply tftypes.ElementKeyInt to block NestingModeSet",
		},
		"WithAttributeName-SetNestedBlocks-WithElementKeyString": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected:    Attribute{},
			expectedErr: "ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to block NestingModeSet",
		},
		"WithAttributeName-SetNestedBlocks-WithElementKeyValue": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "sub_test")),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithAttributeName-SetNestedBlocks-WithElementKeyValue-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other_attr": {
						Type:     types.BoolType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"other_block": {
						Attributes: map[string]Attribute{
							"sub_test": {
								Type:     types.BoolType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "element")).WithAttributeName("sub_test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-SingleNestedAttributes-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-SingleNestedAttributes-WithElementKeyInt": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: "ElementKeyInt(0) still remains in the path: can't apply tftypes.ElementKeyInt to Attributes",
		},
		"WithAttributeName-SingleNestedAttributes-WithElementKeyString": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected:    Attribute{},
			expectedErr: "ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to Attributes",
		},
		"WithAttributeName-SingleNestedAttributes-WithElementKeyValue": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"other": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						}),
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "sub_test")),
			expected:    Attribute{},
			expectedErr: "ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't apply tftypes.ElementKeyValue to Attributes",
		},
		"WithAttributeName-Object-WithAttributeName": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Type: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"sub_test": types.StringType,
							},
						},
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithAttributeName-WithElementKeyInt-invalid-parent": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: "ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to types.StringType",
		},
		"WithAttributeName-WithElementKeyInt-valid-parent": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithAttributeName-WithElementKeyString-invalid-parent": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("element"),
			expected:    Attribute{},
			expectedErr: "ElementKeyString(\"element\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to types.StringType",
		},
		"WithAttributeName-WithElementKeyString-valid-parent": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Type: types.MapType{
							ElemType: types.StringType,
						},
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("element"),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithAttributeName-WithElementKeyValue-invalid-parent": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "element")),
			expected:    Attribute{},
			expectedErr: "ElementKeyValue(tftypes.String<\"element\">) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to types.StringType",
		},
		"WithAttributeName-WithElementKeyValue-valid-parent": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"other": {
						Type:     types.BoolType,
						Optional: true,
					},
					"test": {
						Type: types.SetType{
							ElemType: types.StringType,
						},
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "element")),
			expected:    Attribute{},
			expectedErr: ErrPathInsideAtomicAttribute.Error(),
		},
		"WithElementKeyInt": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: "ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema",
		},
		"WithElementKeyString": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithElementKeyString("test"),
			expected:    Attribute{},
			expectedErr: "ElementKeyString(\"test\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema",
		},
		"WithElementKeyValue": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:    Attribute{},
			expectedErr: "ElementKeyValue(tftypes.String<\"test\">) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema",
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.schema.AttributeAtPath(tc.path)

			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if err.Error() != tc.expectedErr {
					t.Errorf("Expected error to be %q, got %q", tc.expectedErr, err.Error())
					return
				}
				// got expected error
				return
			}

			if err == nil && tc.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", tc.expectedErr)
				return
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSchemaAttributeType(t *testing.T) {
	testSchema := Schema{
		Attributes: map[string]Attribute{
			"foo": {
				Type:     types.StringType,
				Required: true,
			},
			"bar": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required: true,
			},
			"disks": {
				Attributes: ListNestedAttributes(map[string]Attribute{
					"id": {
						Type:     types.StringType,
						Required: true,
					},
					"delete_with_instance": {
						Type:     types.BoolType,
						Optional: true,
					},
				}),
				Optional: true,
				Computed: true,
			},
			"boot_disk": {
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"id": {
						Type:     types.StringType,
						Required: true,
					},
					"delete_with_instance": {
						Type: types.BoolType,
					},
				}),
			},
		},
		Blocks: map[string]Block{
			"list_nested_blocks": {
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
					"number": {
						Type:     types.NumberType,
						Optional: true,
					},
					"bool": {
						Type:     types.BoolType,
						Computed: true,
					},
					"list": {
						Type:     types.ListType{ElemType: types.StringType},
						Computed: true,
						Optional: true,
					},
				},
				NestingMode: BlockNestingModeList,
			},
			"set_nested_blocks": {
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
					"number": {
						Type:     types.NumberType,
						Optional: true,
					},
					"bool": {
						Type:     types.BoolType,
						Computed: true,
					},
					"list": {
						Type:     types.ListType{ElemType: types.StringType},
						Computed: true,
						Optional: true,
					},
				},
				NestingMode: BlockNestingModeSet,
			},
		},
	}

	expectedType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"foo": types.StringType,
			"bar": types.ListType{
				ElemType: types.StringType,
			},
			"disks": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                   types.StringType,
						"delete_with_instance": types.BoolType,
					},
				},
			},
			"boot_disk": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
			},
			"list_nested_blocks": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"string": types.StringType,
						"number": types.NumberType,
						"bool":   types.BoolType,
						"list":   types.ListType{ElemType: types.StringType},
					},
				},
			},
			"set_nested_blocks": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"string": types.StringType,
						"number": types.NumberType,
						"bool":   types.BoolType,
						"list":   types.ListType{ElemType: types.StringType},
					},
				},
			},
		},
	}

	actualType := testSchema.AttributeType()

	if !expectedType.Equal(actualType) {
		t.Fatalf("types not equal (+wanted, -got): %s", cmp.Diff(expectedType, actualType))
	}
}
