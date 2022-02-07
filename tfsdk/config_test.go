package tfsdk

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
				Name: testtypes.String{String: types.String{Value: ""}, CreatedBy: testtypes.StringTypeWithValidateError{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
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
				Name: testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
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
					tftypes.NewAttributePath().WithAttributeName("name"),
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
					tftypes.NewAttributePath().WithAttributeName("name"),
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
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
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
			expected:      &testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"Computed-Computed-object": {
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
			expected:      &testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.config.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"), tc.target)

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
		path          *tftypes.AttributePath
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
			path:     tftypes.NewAttributePath().WithAttributeName("test"),
			expected: nil,
		},
		"WithAttributeName-nonexistent": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("other"),
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("other"),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"AttributeName(\"other\") still remains in the path: could not find attribute or block \"other\" in schema",
				),
			},
		},
		"WithAttributeName-List-null-WithElementKeyInt": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected: types.String{Null: true},
		},
		"WithAttributeName-List-WithElementKeyInt": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-ListNestedAttributes-null-WithElementKeyInt-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-ListNestedAttributes-null-WithElementKeyInt-WithAttributeName-Object": {
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
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: types.Object{
				Null:      true,
				AttrTypes: map[string]attr.Type{"value": types.StringType},
			},
		},
		"WithAttributeName-ListNestedAttributes-WithElementKeyInt-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-ListNestedBlocks-null-WithElementKeyInt-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-ListNestedBlocks-WithElementKeyInt-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-Map-null-WithElementKeyString": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-Map-WithElementKeyString": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-Map-WithElementKeyString-nonexistent": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("other"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-MapNestedAttributes-null-WithElementKeyInt-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("element").WithAttributeName("sub_test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-MapNestedAttributes-WithElementKeyString-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("element").WithAttributeName("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-Object-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-Set-null-WithElementKeyValue": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "value")),
			expected: types.String{Null: true},
		},
		"WithAttributeName-Set-WithElementKeyValue": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "value")),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-SetNestedAttributes-null-WithElementKeyValue-WithAttributeName": {
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
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"sub_test": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"sub_test": tftypes.NewValue(tftypes.String, "value"),
			})).WithAttributeName("sub_test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-SetNestedAttributes-WithElementKeyValue-WithAttributeName": {
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
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"sub_test": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"sub_test": tftypes.NewValue(tftypes.String, "value"),
			})).WithAttributeName("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-SetNestedBlocks-null-WithElementKeyValue-WithAttributeName": {
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
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"sub_test": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"sub_test": tftypes.NewValue(tftypes.String, "value"),
			})).WithAttributeName("sub_test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-SetNestedBlocks-WithElementKeyValue-WithAttributeName": {
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
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"sub_test": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"sub_test": tftypes.NewValue(tftypes.String, "value"),
			})).WithAttributeName("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-SingleNestedAttributes-null-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-SingleNestedAttributes-WithAttributeName": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-String-null": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-String-unknown": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test"),
			expected: types.String{Unknown: true},
		},
		"WithAttributeName-String-value": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test"),
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
			path:          tftypes.NewAttributePath().WithAttributeName("test"),
			expected:      nil,
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"))},
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
			path:          tftypes.NewAttributePath().WithAttributeName("test"),
			expected:      testtypes.String{String: types.String{Value: "value"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test"))},
		},
		"AttrTypeInt64WithValidateError-nested-missing-in-config": {
			config: Config{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"parent": tftypes.Object{},
					},
				}, map[string]tftypes.Value{
					"parent": tftypes.NewValue(tftypes.Object{}, nil),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"parent": {
							Attributes: SingleNestedAttributes(map[string]Attribute{
								"test": {
									Type:     types.Int64Type,
									Optional: true,
									Computed: true,
								},
							}),
							Computed: true,
							Optional: true,
						},
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("parent").WithAttributeName("test"),
			expected:      types.Int64{Null: true},
			expectedDiags: nil,
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
