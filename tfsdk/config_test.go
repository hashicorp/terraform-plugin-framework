package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
						"AttributeName(\"other\") still remains in the path: could not find attribute \"other\" in schema",
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
			expected: nil,
			// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/150
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"ElementKeyInt(0) still remains in the path: step cannot be applied to this value",
				),
			},
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
			expected: nil,
			// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/150
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"ElementKeyInt(0).AttributeName(\"sub_test\") still remains in the path: step cannot be applied to this value",
				),
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
			expected: nil,
			// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/150
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("sub_test"),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"ElementKeyString(\"sub_test\") still remains in the path: step cannot be applied to this value",
				),
			},
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
			expected: nil,
			// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/150
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("other"),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"ElementKeyString(\"other\") still remains in the path: step cannot be applied to this value",
				),
			},
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
			expected: nil,
			// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/150
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("element").WithAttributeName("sub_test"),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"ElementKeyString(\"element\").AttributeName(\"sub_test\") still remains in the path: step cannot be applied to this value",
				),
			},
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
			expected: nil,
			// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/150
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "value")),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"ElementKeyValue(tftypes.String<\"value\">) still remains in the path: step cannot be applied to this value",
				),
			},
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
			expected: nil,
			// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/150
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"sub_test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"sub_test": tftypes.NewValue(tftypes.String, "value"),
					})).WithAttributeName("sub_test"),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"ElementKeyValue(tftypes.Object[\"sub_test\":tftypes.String]<\"sub_test\":tftypes.String<\"value\">>).AttributeName(\"sub_test\") still remains in the path: step cannot be applied to this value",
				),
			},
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
			expected: nil,
			// TODO: https://github.com/hashicorp/terraform-plugin-framework/issues/150
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("sub_test"),
					"Configuration Read Error",
					"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"AttributeName(\"sub_test\") still remains in the path: step cannot be applied to this value",
				),
			},
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
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			val, diags := tc.config.GetAttribute(context.Background(), tc.path)
			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				for _, d := range diags {
					t.Log(d.Summary())
					t.Log(d.Detail())
				}
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
