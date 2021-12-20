package tfsdk

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attrpath"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	intreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestConfigGet(t *testing.T) {
	t.Parallel()

	type testConfigGetData struct {
		Name types.String `tfsdk:"name"`
	}

	type testCase struct {
		config        Config
		expected      testConfigGetData
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"basic": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			expected: testConfigGetData{
				Name: types.String{Value: "namevalue"},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testConfigGetData

			diags := tc.config.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestConfigGet_testTypes(t *testing.T) {
	t.Parallel()

	type testConfigGetData struct {
		Name testtypes.String `tfsdk:"name"`
	}

	type testCase struct {
		config        Config
		expected      testConfigGetData
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"AttrTypeWithValidateError": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			expected: testConfigGetData{
				Name: testtypes.String{Str: types.String{Value: ""}, CreatedBy: testtypes.StringTypeWithValidateError{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(attrpath.New().Attribute("name"))},
		},
		"AttrTypeWithValidateWarning": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			expected: testConfigGetData{
				Name: testtypes.String{Str: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(attrpath.New().Attribute("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testConfigGetData

			diags := tc.config.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestConfigGetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		config        Config
		target        interface{}
		expected      interface{}
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"string": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(string),
			expected: newStringPointer("namevalue"),
		},
		"*string": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(*string),
			expected: newStringPointerPointer("namevalue"),
		},
		"types.String": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(types.String),
			expected: &types.String{Value: "namevalue"},
		},
		"incompatible-target": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(testtypes.String),
			expected: new(testtypes.String),
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					attrpath.New().Attribute("name"),
					intreflect.DiagNewAttributeValueIntoWrongType{
						ValType:    reflect.TypeOf(types.String{Value: "namevalue"}),
						TargetType: reflect.TypeOf(testtypes.String{}),
						SchemaType: types.StringType,
					},
				),
			},
		},
		"incompatible-type": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(bool),
			expected: new(bool),
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					attrpath.New().Attribute("name"),
					intreflect.DiagIntoIncompatibleType{
						Val:        tftypes.NewValue(tftypes.String, "namevalue"),
						TargetType: reflect.TypeOf(false),
						Err:        fmt.Errorf("can't unmarshal %s into *%T, expected boolean", tftypes.String, false),
					},
				),
			},
		},
		"AttrTypeWithValidateError": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			target:        new(testtypes.String),
			expected:      new(testtypes.String),
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(attrpath.New().Attribute("name"))},
		},
		"AttrTypeWithValidateWarning": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			target:        new(testtypes.String),
			expected:      &testtypes.String{Str: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(attrpath.New().Attribute("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.config.GetAttribute(context.Background(), attrpath.New().Attribute("name"), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected, cmp.Transformer("testtypes", func(in *testtypes.String) testtypes.String { return *in }), cmp.Transformer("types", func(in *types.String) types.String { return *in })); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestConfigGetAttributeValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		config        Config
		path          attrpath.Path
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"empty": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test"),
			expected: nil,
		},
		"Attribute-nonexistent": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "value"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path:     attrpath.New().Attribute("other"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					attrpath.New().Attribute("other"),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"AttributeName(\"other\") still remains in the path: could not find attribute or block \"other\" in schema",
				),
			},
		},
		"Attribute-List-null-ElementPos": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test").ElementPos(0),
			expected: types.String{Null: true},
		},
		"Attribute-List-ElementPos": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test").ElementPos(0),
			expected: types.String{Value: "value"},
		},
		"Attribute-ListNestedAttributes-null-ElementPos-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: ListNestedAttributes(map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}, ListNestedAttributesOptions{}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     attrpath.New().Attribute("test").ElementPos(0).Attribute("sub_test"),
			expected: types.String{Null: true},
		},
		"Attribute-ListNestedAttributes-null-ElementPos-Attribute-Object": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: ListNestedAttributes(map[string]Attribute{
								"sub_test": {
									Attributes: SingleNestedAttributes(map[string]Attribute{
										"value": {
											Type:     types.StringType,
											Optional: true,
										},
									}),
									Optional: true,
								},
							}, ListNestedAttributesOptions{}),
							Optional: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: attrpath.New().Attribute("test").ElementPos(0).Attribute("sub_test"),
			expected: types.Object{
				Null:      true,
				AttrTypes: map[string]attr.Type{"value": types.StringType},
			},
		},
		"Attribute-ListNestedAttributes-ElementPos-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: ListNestedAttributes(map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}, ListNestedAttributesOptions{}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     attrpath.New().Attribute("test").ElementPos(0).Attribute("sub_test"),
			expected: types.String{Value: "value"},
		},
		"Attribute-ListNestedBlocks-null-ElementPos-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
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
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: BlockNestingModeList,
						},
					},
				},
			},
			path:     attrpath.New().Attribute("test").ElementPos(0).Attribute("sub_test"),
			expected: types.String{Null: true},
		},
		"Attribute-ListNestedBlocks-ElementPos-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
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
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: BlockNestingModeList,
						},
					},
				},
			},
			path:     attrpath.New().Attribute("test").ElementPos(0).Attribute("sub_test"),
			expected: types.String{Value: "value"},
		},
		"Attribute-Map-null-ElementKey": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test").ElementKey("sub_test"),
			expected: types.String{Null: true},
		},
		"Attribute-Map-ElementKey": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test").ElementKey("sub_test"),
			expected: types.String{Value: "value"},
		},
		"Attribute-Map-ElementKey-nonexistent": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test").ElementKey("other"),
			expected: types.String{Null: true},
		},
		"Attribute-MapNestedAttributes-null-ElementPos-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: MapNestedAttributes(map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}, MapNestedAttributesOptions{}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     attrpath.New().Attribute("test").ElementKey("element").Attribute("sub_test"),
			expected: types.String{Null: true},
		},
		"Attribute-MapNestedAttributes-ElementKey-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: MapNestedAttributes(map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}, MapNestedAttributesOptions{}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path:     attrpath.New().Attribute("test").ElementKey("element").Attribute("sub_test"),
			expected: types.String{Value: "value"},
		},
		"Attribute-Object-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test").Attribute("sub_test"),
			expected: types.String{Value: "value"},
		},
		"Attribute-Set-null-WithElementKeyValue": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test"), // TODO: restore when we've fixed set values: .WithElementKeyValue(tftypes.NewValue(tftypes.String, "value")),
			expected: types.String{Null: true},
		},
		"Attribute-Set-WithElementKeyValue": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test"), // TODO: restore when we've fixed set values: .WithElementKeyValue(tftypes.NewValue(tftypes.String, "value")),
			expected: types.String{Value: "value"},
		},
		"Attribute-SetNestedAttributes-null-WithElementKeyValue-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: SetNestedAttributes(map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}, SetNestedAttributesOptions{}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: attrpath.New().Attribute("test"), /* TODO: restore when we've fixed set values.WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"sub_test": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"sub_test": tftypes.NewValue(tftypes.String, "value"),
			})).Attribute("sub_test"),*/
			expected: types.String{Null: true},
		},
		"Attribute-SetNestedAttributes-WithElementKeyValue-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: SetNestedAttributes(map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							}, SetNestedAttributesOptions{}),
							Required: true,
						},
						"other": {
							Type:     types.BoolType,
							Optional: true,
						},
					},
				},
			},
			path: attrpath.New().Attribute("test"), /* TODO: restore when we've fixed set values: .WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"sub_test": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"sub_test": tftypes.NewValue(tftypes.String, "value"),
			})).Attribute("sub_test"),*/
			expected: types.String{Value: "value"},
		},
		"Attribute-SetNestedBlocks-null-WithElementKeyValue-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
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
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: BlockNestingModeSet,
						},
					},
				},
			},
			path: attrpath.New().Attribute("test"), /* TODO: restore when we've fixed set values: .WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"sub_test": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"sub_test": tftypes.NewValue(tftypes.String, "value"),
			})).Attribute("sub_test"),*/
			expected: types.String{Null: true},
		},
		"Attribute-SetNestedBlocks-WithElementKeyValue-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
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
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: BlockNestingModeSet,
						},
					},
				},
			},
			path: attrpath.New().Attribute("test"), /* TODO: restore when we've fixed set values: .WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"sub_test": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"sub_test": tftypes.NewValue(tftypes.String, "value"),
			})).Attribute("sub_test"),*/
			expected: types.String{Value: "value"},
		},
		"Attribute-SingleNestedAttributes-null-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: SingleNestedAttributes(map[string]Attribute{
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
			path:     attrpath.New().Attribute("test").Attribute("sub_test"),
			expected: types.String{Null: true},
		},
		"Attribute-SingleNestedAttributes-Attribute": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Attributes: SingleNestedAttributes(map[string]Attribute{
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
			path:     attrpath.New().Attribute("test").Attribute("sub_test"),
			expected: types.String{Value: "value"},
		},
		"Attribute-String-null": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, nil),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test"),
			expected: types.String{Null: true},
		},
		"Attribute-String-unknown": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test"),
			expected: types.String{Unknown: true},
		},
		"Attribute-String-value": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:     attrpath.New().Attribute("test"),
			expected: types.String{Value: "value"},
		},
		"AttrTypeWithValidateError": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:          attrpath.New().Attribute("test"),
			expected:      nil,
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(attrpath.New().Attribute("test"))},
		},
		"AttrTypeWithValidateWarning": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "value"),
					"other": tftypes.NewValue(tftypes.Bool, nil),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
			path:          attrpath.New().Attribute("test"),
			expected:      testtypes.String{Str: types.String{Value: "value"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(attrpath.New().Attribute("test"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			val, diags := tc.config.getAttributeValue(context.Background(), tc.path)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
