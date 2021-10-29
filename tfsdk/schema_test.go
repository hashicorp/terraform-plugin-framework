package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
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
						}, ListNestedAttributesOptions{}),
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
						}, ListNestedAttributesOptions{}),
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
						}, ListNestedAttributesOptions{}),
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
						}, ListNestedAttributesOptions{}),
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
						}, ListNestedAttributesOptions{}),
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						NestingMode: NestingModeList,
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
						}, MapNestedAttributesOptions{}),
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
						}, MapNestedAttributesOptions{}),
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
						}, MapNestedAttributesOptions{}),
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
						}, MapNestedAttributesOptions{}),
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
						}, MapNestedAttributesOptions{}),
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
						}, SetNestedAttributesOptions{}),
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
						}, SetNestedAttributesOptions{}),
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
						}, SetNestedAttributesOptions{}),
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
						}, SetNestedAttributesOptions{}),
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
						}, SetNestedAttributesOptions{}),
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
						NestingMode: NestingModeSet,
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
				}, ListNestedAttributesOptions{}),
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
				NestingMode: NestingModeList,
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
				NestingMode: NestingModeSet,
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

func TestSchemaTfprotov6Schema(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Schema
		expected    *tfprotov6.Schema
		expectedErr string
	}

	tests := map[string]testCase{
		"empty-val": {
			input:       Schema{},
			expectedErr: "must have at least one attribute or block in the schema",
		},
		"basic-attrs": {
			input: Schema{
				Version: 1,
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
				},
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bool",
							Type:     tftypes.Bool,
							Computed: true,
						},
						{
							Name:     "number",
							Type:     tftypes.Number,
							Optional: true,
						},
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
				},
			},
		},
		"complex-attrs": {
			input: Schema{
				Version: 2,
				Attributes: map[string]Attribute{
					"list": {
						Type:     types.ListType{ElemType: types.StringType},
						Required: true,
					},
					"object": {
						Type: types.ObjectType{AttrTypes: map[string]attr.Type{
							"string": types.StringType,
							"number": types.NumberType,
							"bool":   types.BoolType,
						}},
						Optional: true,
					},
					"map": {
						Type:     types.MapType{ElemType: types.NumberType},
						Computed: true,
					},
					"set": {
						Type:     types.SetType{ElemType: types.StringType},
						Required: true,
					},
					// TODO: add tuple support when it lands
				},
			},
			expected: &tfprotov6.Schema{
				Version: 2,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "list",
							Type:     tftypes.List{ElementType: tftypes.String},
							Required: true,
						},
						{
							Name:     "map",
							Type:     tftypes.Map{ElementType: tftypes.Number},
							Computed: true,
						},
						{
							Name: "object",
							Type: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
								"string": tftypes.String,
								"number": tftypes.Number,
								"bool":   tftypes.Bool,
							}},
							Optional: true,
						},
						{
							Name:     "set",
							Type:     tftypes.Set{ElementType: tftypes.String},
							Required: true,
						},
					},
				},
			},
		},
		"nested-attrs": {
			input: Schema{
				Version: 3,
				Attributes: map[string]Attribute{
					"single": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
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
						}),
						Required: true,
					},
					"list": {
						Attributes: ListNestedAttributes(map[string]Attribute{
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
						}, ListNestedAttributesOptions{}),
						Optional: true,
					},
					"set": {
						Attributes: SetNestedAttributes(map[string]Attribute{
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
						}, SetNestedAttributesOptions{}),
						Computed: true,
					},
					"map": {
						Attributes: MapNestedAttributes(map[string]Attribute{
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
						}, MapNestedAttributesOptions{}),
						Optional: true,
						Computed: true,
					},
				},
			},
			expected: &tfprotov6.Schema{
				Version: 3,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name: "list",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeList,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Optional: true,
						},
						{
							Name: "map",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeMap,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Optional: true,
							Computed: true,
						},
						{
							Name: "set",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeSet,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Computed: true,
						},
						{
							Name: "single",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeSingle,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
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
		"nested-blocks": {
			input: Schema{
				Version: 3,
				Blocks: map[string]Block{
					"list": {
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
						NestingMode: NestingModeList,
					},
					"set": {
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
						NestingMode: NestingModeSet,
					},
				},
			},
			expected: &tfprotov6.Schema{
				Version: 3,
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Computed: true,
										Name:     "bool",
										Type:     tftypes.Bool,
									},
									{
										Computed: true,
										Name:     "list",
										Optional: true,
										Type:     tftypes.List{ElementType: tftypes.String},
									},
									{
										Name:     "number",
										Optional: true,
										Type:     tftypes.Number,
									},
									{
										Name:     "string",
										Required: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
							TypeName: "list",
						},
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Computed: true,
										Name:     "bool",
										Type:     tftypes.Bool,
									},
									{
										Computed: true,
										Name:     "list",
										Optional: true,
										Type:     tftypes.List{ElementType: tftypes.String},
									},
									{
										Name:     "number",
										Optional: true,
										Type:     tftypes.Number,
									},
									{
										Name:     "string",
										Required: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
							TypeName: "set",
						},
					},
				},
			},
		},
		"markdown-description": {
			input: Schema{
				Version: 1,
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
				},
				MarkdownDescription: "a test resource",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Description:     "a test resource",
					DescriptionKind: tfprotov6.StringKindMarkdown,
				},
			},
		},
		"plaintext-description": {
			input: Schema{
				Version: 1,
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
				},
				Description: "a test resource",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Description:     "a test resource",
					DescriptionKind: tfprotov6.StringKindPlain,
				},
			},
		},
		"deprecated": {
			input: Schema{
				Version: 1,
				Attributes: map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
				},
				DeprecationMessage: "deprecated, use other_resource instead",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Deprecated: true,
				},
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.input.tfprotov6Schema(context.Background())
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
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

func TestSchemaValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req  ValidateSchemaRequest
		resp ValidateSchemaResponse
	}{
		"no-validation": {
			req: ValidateSchemaRequest{
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{},
		},
		"deprecation-message": {
			req: ValidateSchemaRequest{
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
							},
						},
						DeprecationMessage: "Use something else instead.",
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic(
						"Deprecated",
						"Use something else instead.",
					),
				},
			},
		},
		"warnings": {
			req: ValidateSchemaRequest{
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testWarningAttributeValidator{},
								},
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testWarningAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: diag.Diagnostics{
					testWarningDiagnostic1,
					testWarningDiagnostic2,
				},
			},
		},
		"errors": {
			req: ValidateSchemaRequest{
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"attr1": tftypes.String,
							"attr2": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"attr1": tftypes.NewValue(tftypes.String, "attr1value"),
						"attr2": tftypes.NewValue(tftypes.String, "attr2value"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"attr1": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testErrorAttributeValidator{},
								},
							},
							"attr2": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testErrorAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: ValidateSchemaResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
					testErrorDiagnostic2,
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got ValidateSchemaResponse
			tc.req.Config.Schema.validate(context.Background(), tc.req, &got)

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}
