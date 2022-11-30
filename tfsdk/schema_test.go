package tfsdk

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSchemaAttributeAtPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        Schema
		path          path.Path
		expected      fwschema.Attribute
		expectedDiags diag.Diagnostics
	}{
		"empty-root": {
			schema:   Schema{},
			path:     path.Empty(),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: \n"+
						"Original Error: got unexpected type tfsdk.Schema",
				),
			},
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
			path:     path.Empty(),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: \n"+
						"Original Error: got unexpected type tfsdk.Schema",
				),
			},
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
			path: path.Root("test"),
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
			path:     path.Root("test").AtName("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test.sub_test\n"+
						"Original Error: AttributeName(\"sub_test\") still remains in the path: can't apply tftypes.AttributeName to ListNestedAttributes",
				),
			},
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
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path: path.Root("test").AtListIndex(0).AtName("sub_test"),
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
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"sub_test\"]\n"+
						"Original Error: ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to ListNestedAttributes",
				),
			},
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
			path:     path.Root("test").AtSetValue(types.StringValue("sub_test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("sub_test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"sub_test\")]\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't apply tftypes.ElementKeyValue to ListNestedAttributes",
				),
			},
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
			path:     path.Root("test").AtName("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test.sub_test\n"+
						"Original Error: AttributeName(\"sub_test\") still remains in the path: can't apply tftypes.AttributeName to block NestingModeList",
				),
			},
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
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path: path.Root("test").AtListIndex(0).AtName("sub_test"),
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
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"sub_test\"]\n"+
						"Original Error: ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to block NestingModeList",
				),
			},
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
			path:     path.Root("test").AtSetValue(types.StringValue("sub_test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("sub_test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"sub_test\")]\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't apply tftypes.ElementKeyValue to block NestingModeList",
				),
			},
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
			path:     path.Root("test").AtName("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test.sub_test\n"+
						"Original Error: AttributeName(\"sub_test\") still remains in the path: can't use tftypes.AttributeName on maps",
				),
			},
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
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: can't use tftypes.ElementKeyInt on maps",
				),
			},
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
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"sub_test\"]\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path: path.Root("test").AtMapKey("element").AtName("sub_test"),
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
			path:     path.Root("test").AtSetValue(types.StringValue("sub_test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("sub_test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"sub_test\")]\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't use tftypes.ElementKeyValue on maps",
				),
			},
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
			path:     path.Root("test").AtName("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test.sub_test\n"+
						"Original Error: AttributeName(\"sub_test\") still remains in the path: can't use tftypes.AttributeName on sets",
				),
			},
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
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: can't use tftypes.ElementKeyInt on sets",
				),
			},
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
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"sub_test\"]\n"+
						"Original Error: ElementKeyString(\"sub_test\") still remains in the path: can't use tftypes.ElementKeyString on sets",
				),
			},
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
			path:     path.Root("test").AtSetValue(types.StringValue("sub_test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("sub_test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"sub_test\")]\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path: path.Root("test").AtSetValue(types.StringValue("element")).AtName("sub_test"),
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
			path:     path.Root("test").AtName("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test.sub_test\n"+
						"Original Error: AttributeName(\"sub_test\") still remains in the path: can't apply tftypes.AttributeName to block NestingModeSet",
				),
			},
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
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: can't apply tftypes.ElementKeyInt to block NestingModeSet",
				),
			},
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
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"sub_test\"]\n"+
						"Original Error: ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to block NestingModeSet",
				),
			},
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
			path:     path.Root("test").AtSetValue(types.StringValue("sub_test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("sub_test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"sub_test\")]\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path: path.Root("test").AtSetValue(types.StringValue("element")).AtName("sub_test"),
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
			path: path.Root("test").AtName("sub_test"),
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
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: can't apply tftypes.ElementKeyInt to Attributes",
				),
			},
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
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"sub_test\"]\n"+
						"Original Error: ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to Attributes",
				),
			},
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
			path:     path.Root("test").AtSetValue(types.StringValue("sub_test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("sub_test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"sub_test\")]\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't apply tftypes.ElementKeyValue to Attributes",
				),
			},
		},
		"WithAttributeName-SingleBlock-WithAttributeName": {
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
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_nested_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path: path.Root("test").AtName("sub_test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-SingleBlock-WithElementKeyInt": {
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
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_nested_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: can't apply tftypes.ElementKeyInt to block NestingModeSingle",
				),
			},
		},
		"WithAttributeName-SingleBlock-WithElementKeyString": {
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
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_nested_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path:     path.Root("test").AtMapKey("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"sub_test\"]\n"+
						"Original Error: ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to block NestingModeSingle",
				),
			},
		},
		"WithAttributeName-SingleBlock-WithElementKeyValue": {
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
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_nested_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path:     path.Root("test").AtSetValue(types.StringValue("sub_test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("sub_test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"sub_test\")]\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't apply tftypes.ElementKeyValue to block NestingModeSingle",
				),
			},
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
			path:     path.Root("test").AtName("sub_test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtName("sub_test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test.sub_test\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.StringType",
				),
			},
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
			path:     path.Root("test").AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[0]\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path:     path.Root("test").AtMapKey("element"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("element"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"element\"]\n"+
						"Original Error: ElementKeyString(\"element\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.StringType",
				),
			},
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
			path:     path.Root("test").AtMapKey("element"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("element"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[\"element\"]\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path:     path.Root("test").AtSetValue(types.StringValue("element")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("element")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"element\")]\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"element\">) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.StringType",
				),
			},
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
			path:     path.Root("test").AtSetValue(types.StringValue("element")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtSetValue(types.StringValue("element")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: test[Value(\"element\")]\n"+
						"Original Error: "+ErrPathInsideAtomicAttribute.Error(),
				),
			},
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
			path:     path.Empty().AtListIndex(0),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema",
				),
			},
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
			path:     path.Empty().AtMapKey("test"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtMapKey("test"),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [\"test\"]\n"+
						"Original Error: ElementKeyString(\"test\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema",
				),
			},
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
			path:     path.Empty().AtSetValue(types.StringValue("test")),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtSetValue(types.StringValue("test")),
					"Invalid Schema Path",
					"When attempting to get the framework attribute associated with a schema path, an unexpected error was returned. "+
						"This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [Value(\"test\")]\n"+
						"Original Error: ElementKeyValue(tftypes.String<\"test\">) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema",
				),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := tc.schema.AttributeAtPath(context.Background(), tc.path)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSchemaAttributeAtTerraformPath(t *testing.T) {
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
		"WithAttributeName-SingleBlock-WithAttributeName": {
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
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_nested_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected: Attribute{
				Type:     types.StringType,
				Required: true,
			},
		},
		"WithAttributeName-SingleBlock-WithElementKeyInt": {
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
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_nested_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected:    Attribute{},
			expectedErr: "ElementKeyInt(0) still remains in the path: can't apply tftypes.ElementKeyInt to block NestingModeSingle",
		},
		"WithAttributeName-SingleBlock-WithElementKeyString": {
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
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_nested_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected:    Attribute{},
			expectedErr: "ElementKeyString(\"sub_test\") still remains in the path: can't apply tftypes.ElementKeyString to block NestingModeSingle",
		},
		"WithAttributeName-SingleBlock-WithElementKeyValue": {
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
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
					"test": {
						Attributes: map[string]Attribute{
							"other_nested_attr": {
								Type:     types.BoolType,
								Optional: true,
							},
							"sub_test": {
								Type:     types.StringType,
								Required: true,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path:        tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "sub_test")),
			expected:    Attribute{},
			expectedErr: "ElementKeyValue(tftypes.String<\"sub_test\">) still remains in the path: can't apply tftypes.ElementKeyValue to block NestingModeSingle",
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
			expectedErr: "ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.StringType",
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
			expectedErr: "ElementKeyString(\"element\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.StringType",
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
			expectedErr: "ElementKeyValue(tftypes.String<\"element\">) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.StringType",
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

			got, err := tc.schema.AttributeAtTerraformPath(context.Background(), tc.path)

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

func TestSchemaType(t *testing.T) {
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
			"single_nested_block": {
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
				NestingMode: BlockNestingModeSingle,
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
			"single_nested_block": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"string": types.StringType,
					"number": types.NumberType,
					"bool":   types.BoolType,
					"list":   types.ListType{ElemType: types.StringType},
				},
			},
		},
	}

	actualType := testSchema.Type()

	if !expectedType.Equal(actualType) {
		t.Fatalf("types not equal (+wanted, -got): %s", cmp.Diff(expectedType, actualType))
	}
}

func TestSchemaTypeAtPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        Schema
		path          path.Path
		expected      attr.Type
		expectedDiags diag.Diagnostics
	}{
		"empty-schema-empty-path": {
			schema:   Schema{},
			path:     path.Empty(),
			expected: types.ObjectType{},
		},
		"empty-path": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"bool": {
						Required: true,
						Type:     types.BoolType,
					},
					"string": {
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path: path.Empty(),
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bool":   types.BoolType,
					"string": types.StringType,
				},
			},
		},
		"AttributeName-Attribute": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"bool": {
						Required: true,
						Type:     types.BoolType,
					},
					"string": {
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path:     path.Root("string"),
			expected: types.StringType,
		},
		"AttributeName-Block": {
			schema: Schema{
				Blocks: map[string]Block{
					"list_block": {
						Attributes: map[string]Attribute{
							"list_block_nested": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeList,
					},
					"set_block": {
						Attributes: map[string]Attribute{
							"set_block_nested": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
					"single_block": {
						Attributes: map[string]Attribute{
							"single_block_nested": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path: path.Root("list_block"),
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"list_block_nested": types.StringType,
					},
				},
			},
		},
		"AttributeName-non-existent": {
			schema: Schema{},
			path:   path.Root("non-existent"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("non-existent"),
					"Invalid Schema Path",
					"When attempting to get the framework type associated with a schema path, an unexpected error was returned. This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: non-existent\n"+
						"Original Error: AttributeName(\"non-existent\") still remains in the path: could not find attribute or block \"non-existent\" in schema",
				),
			},
		},
		"ElementKeyInt": {
			schema: Schema{},
			path:   path.Empty().AtListIndex(0),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtListIndex(0),
					"Invalid Schema Path",
					"When attempting to get the framework type associated with a schema path, an unexpected error was returned. This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [0]\n"+
						"Original Error: ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema",
				),
			},
		},
		"ElementKeyString": {
			schema: Schema{},
			path:   path.Empty().AtMapKey("invalid"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtMapKey("invalid"),
					"Invalid Schema Path",
					"When attempting to get the framework type associated with a schema path, an unexpected error was returned. This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [\"invalid\"]\n"+
						"Original Error: ElementKeyString(\"invalid\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema",
				),
			},
		},
		"ElementKeyValue": {
			schema: Schema{},
			path:   path.Empty().AtSetValue(types.StringNull()),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtSetValue(types.StringNull()),
					"Invalid Schema Path",
					"When attempting to get the framework type associated with a schema path, an unexpected error was returned. This is always an issue with the provider. Please report this to the provider developers.\n\n"+
						"Path: [Value(<null>)]\n"+
						"Original Error: ElementKeyValue(tftypes.String<null>) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := testCase.schema.TypeAtPath(context.Background(), testCase.path)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSchemaTypeAtTerraformPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        Schema
		path          *tftypes.AttributePath
		expected      attr.Type
		expectedError error
	}{
		"empty-schema-nil-path": {
			schema:   Schema{},
			path:     nil,
			expected: types.ObjectType{},
		},
		"empty-schema-empty-path": {
			schema:   Schema{},
			path:     tftypes.NewAttributePath(),
			expected: types.ObjectType{},
		},
		"nil-path": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"bool": {
						Required: true,
						Type:     types.BoolType,
					},
					"string": {
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path: nil,
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bool":   types.BoolType,
					"string": types.StringType,
				},
			},
		},
		"empty-path": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"bool": {
						Required: true,
						Type:     types.BoolType,
					},
					"string": {
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path: tftypes.NewAttributePath(),
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"bool":   types.BoolType,
					"string": types.StringType,
				},
			},
		},
		"AttributeName-Attribute": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"bool": {
						Required: true,
						Type:     types.BoolType,
					},
					"string": {
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("string"),
			expected: types.StringType,
		},
		"AttributeName-Block": {
			schema: Schema{
				Blocks: map[string]Block{
					"list_block": {
						Attributes: map[string]Attribute{
							"list_block_nested": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeList,
					},
					"set_block": {
						Attributes: map[string]Attribute{
							"set_block_nested": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
					"single_block": {
						Attributes: map[string]Attribute{
							"single_block_nested": {
								Required: true,
								Type:     types.StringType,
							},
						},
						NestingMode: BlockNestingModeSingle,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("list_block"),
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"list_block_nested": types.StringType,
					},
				},
			},
		},
		"AttributeName-non-existent": {
			schema:        Schema{},
			path:          tftypes.NewAttributePath().WithAttributeName("non-existent"),
			expectedError: fmt.Errorf("AttributeName(\"non-existent\") still remains in the path: could not find attribute or block \"non-existent\" in schema"),
		},
		"ElementKeyInt": {
			schema:        Schema{},
			path:          tftypes.NewAttributePath().WithElementKeyInt(0),
			expectedError: fmt.Errorf("ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to schema"),
		},
		"ElementKeyString": {
			schema:        Schema{},
			path:          tftypes.NewAttributePath().WithElementKeyString("invalid"),
			expectedError: fmt.Errorf("ElementKeyString(\"invalid\") still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyString to schema"),
		},
		"ElementKeyValue": {
			schema:        Schema{},
			path:          tftypes.NewAttributePath().WithElementKeyValue(tftypes.NewValue(tftypes.String, nil)),
			expectedError: fmt.Errorf("ElementKeyValue(tftypes.String<null>) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyValue to schema"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.schema.TypeAtTerraformPath(context.Background(), testCase.path)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
