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

func TestPlanGet(t *testing.T) {
	t.Parallel()

	type testPlanGetData struct {
		Name types.String `tfsdk:"name"`
	}

	type testCase struct {
		plan          Plan
		expected      testPlanGetData
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"basic": {
			plan: Plan{
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
			expected: testPlanGetData{
				Name: types.String{Value: "namevalue"},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testPlanGetData

			diags := tc.plan.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanGet_testTypes(t *testing.T) {
	t.Parallel()

	type testPlanGetDataTestTypes struct {
		Name testtypes.String `tfsdk:"name"`
	}

	type testCase struct {
		plan          Plan
		expected      testPlanGetDataTestTypes
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"AttrTypeWithValidateError": {
			plan: Plan{
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
			expected: testPlanGetDataTestTypes{
				Name: testtypes.String{String: types.String{Value: ""}, CreatedBy: testtypes.StringTypeWithValidateError{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
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
			expected: testPlanGetDataTestTypes{
				Name: testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testPlanGetDataTestTypes

			diags := tc.plan.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanGetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		target        interface{}
		expected      interface{}
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"string": {
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("name"),
					"Value Conversion Error",
					intreflect.DiagNewAttributeValueIntoWrongType{
						ValType:    reflect.TypeOf(types.String{Value: "namevalue"}),
						TargetType: reflect.TypeOf(testtypes.String{}),
						AttrPath:   tftypes.NewAttributePath().WithAttributeName("name"),
						SchemaType: types.StringType,
					}.Detail(),
				),
			},
		},
		"incompatible-type": {
			plan: Plan{
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
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("name"),
					"Value Conversion Error",
					intreflect.DiagIntoIncompatibleType{
						Val:        tftypes.NewValue(tftypes.String, "namevalue"),
						TargetType: reflect.TypeOf(false),
						Err:        fmt.Errorf("can't unmarshal %s into *%T, expected boolean", tftypes.String, false),
						AttrPath:   tftypes.NewAttributePath().WithAttributeName("name"),
					}.Detail(),
				),
			},
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
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
			plan: Plan{
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

			diags := tc.plan.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected, cmp.Transformer("testtypes", func(in *testtypes.String) testtypes.String { return *in }), cmp.Transformer("types", func(in *types.String) types.String { return *in })); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanGetAttributeValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		path          *tftypes.AttributePath
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"empty": {
			plan: Plan{
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
			plan: Plan{
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
					"Plan Read Error",
					"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"error getting attribute type in schema: AttributeName(\"other\") still remains in the path: could not find attribute or block \"other\" in schema",
				),
			},
		},
		"WithAttributeName-List-null-WithElementKeyInt": {
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
		"WithAttributeName-ListNestedAttributes-WithElementKeyInt-WithAttributeName": {
			plan: Plan{
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
			plan: Plan{
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
							NestingMode: NestingModeList,
						},
						"test": {
							Attributes: map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: NestingModeList,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: types.String{Null: true},
		},
		"WithAttributeName-ListNestedBlocks-WithElementKeyInt-WithAttributeName": {
			plan: Plan{
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
							NestingMode: NestingModeList,
						},
						"test": {
							Attributes: map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: NestingModeList,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("sub_test"),
			expected: types.String{Value: "value"},
		},
		"WithAttributeName-Map-null-WithElementKeyString": {
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
							NestingMode: NestingModeSet,
						},
						"test": {
							Attributes: map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: NestingModeSet,
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
			plan: Plan{
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
							NestingMode: NestingModeSet,
						},
						"test": {
							Attributes: map[string]Attribute{
								"sub_test": {
									Type:     types.StringType,
									Required: true,
								},
							},
							NestingMode: NestingModeSet,
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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
			plan: Plan{
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

			val, diags := tc.plan.getAttributeValue(context.Background(), tc.path)
			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanPathExists(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		path          *tftypes.AttributePath
		expected      bool
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"empty-path": {
			plan:     Plan{},
			path:     tftypes.NewAttributePath().WithAttributeName("test"),
			expected: false,
		},
		"empty-root": {
			plan:     Plan{},
			path:     tftypes.NewAttributePath(),
			expected: true,
		},
		"root": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     tftypes.NewAttributePath(),
			expected: true,
		},
		"WithAttributeName": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     tftypes.NewAttributePath().WithAttributeName("test"),
			expected: true,
		},
		"WithAttributeName-mismatch": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			expected: false,
		},
		"WithAttributeName.WithAttributeName": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("nested"),
			expected: true,
		},
		"WithAttributeName.WithAttributeName-mismatch-child": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("other"),
			expected: false,
		},
		"WithAttributeName.WithAttributeName-mismatch-parent": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithAttributeName("other"),
			expected: false,
		},
		"WithAttributeName.WithElementKeyInt": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected: true,
		},
		"WithAttributeName.WithElementKeyInt-mismatch-child": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(1),
			expected: false,
		},
		"WithAttributeName.WithElementKeyInt-mismatch-parent": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			expected: false,
		},
		"WithAttributeName.WithElementKeyString": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("key"),
			expected: true,
		},
		"WithAttributeName.WithElementKeyString-mismatch-child": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("other"),
			expected: false,
		},
		"WithAttributeName.WithElementKeyString-mismatch-parent": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("other"),
			expected: false,
		},
		"WithAttributeName.WithElementKeyValue": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "testvalue")),
			expected: true,
		},
		"WithAttributeName.WithElementKeyValue-mismatch-child": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "othervalue")),
			expected: false,
		},
		"WithAttributeName.WithElementKeyValue-mismatch-parent": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "testvalue"),
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
			path:     tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "othervalue")),
			expected: false,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := tc.plan.pathExists(context.Background(), tc.path)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"write": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "oldvalue"),
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
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"overwrite": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected:      tftypes.Value{},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.plan.Set(context.Background(), tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.plan.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanSetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		path          *tftypes.AttributePath
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"add-List-Element-append": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(1),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk0"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"add-List-Element-append-length-error": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(2),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk0"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("disks"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 3 as list currently has 1 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"add-List-Element-first": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"add-List-Element-first-length-error": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(1),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, nil),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("disks"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 2 as list currently has 0 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"add-Map-Element-append": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"key1": tftypes.NewValue(tftypes.String, "key1value"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
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
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("key2"),
			val:  "key2value",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key1": tftypes.NewValue(tftypes.String, "key1value"),
					"key2": tftypes.NewValue(tftypes.String, "key2value"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"add-Map-Element-first": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
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
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("key"),
			val:  "keyvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key": tftypes.NewValue(tftypes.String, "keyvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"add-Set-Element-append": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: SetNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, SetNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":                   tftypes.String,
					"delete_with_instance": tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
			})),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk0"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"add-Set-Element-first": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: SetNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, SetNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":                   tftypes.String,
					"delete_with_instance": tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
			})),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Bool": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.Bool,
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.Bool, true),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type:     types.BoolType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			val:  false,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.Bool,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.Bool, false),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-List": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags":  tftypes.List{ElementType: tftypes.String},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"tags": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"tags": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("tags"),
			val:  []string{"one", "two"},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags":  tftypes.List{ElementType: tftypes.String},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-List-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk1"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(1),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk0"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Map": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"originalkey": tftypes.NewValue(tftypes.String, "originalvalue"),
						"otherkey":    tftypes.NewValue(tftypes.String, "othervalue"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
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
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			val: map[string]string{
				"newkey": "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"newkey": tftypes.NewValue(tftypes.String, "newvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Map-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"key":   tftypes.NewValue(tftypes.String, "originalvalue"),
						"other": tftypes.NewValue(tftypes.String, "should be untouched"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
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
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("key"),
			val:  "newvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key":   tftypes.NewValue(tftypes.String, "newvalue"),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Number": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.Number,
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.Number, 1),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type:     types.NumberType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			val:  2,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.Number,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.Number, 2),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Object": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
								"other":     tftypes.String,
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"scratch_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
							"other":     tftypes.String,
						},
					}, map[string]tftypes.Value{
						"interface": tftypes.NewValue(tftypes.String, "SCSI"),
						"other":     tftypes.NewValue(tftypes.String, "originalvalue"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"scratch_disk": {
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"interface": types.StringType,
									"other":     types.StringType,
								},
							},
							Optional: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("scratch_disk"),
			val: struct {
				Interface string `tfsdk:"interface"`
				Other     string `tfsdk:"other"`
			}{
				Interface: "NVME",
				Other:     "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"scratch_disk": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
							"other":     tftypes.String,
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"scratch_disk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"interface": tftypes.String,
						"other":     tftypes.String,
					},
				}, map[string]tftypes.Value{
					"interface": tftypes.NewValue(tftypes.String, "NVME"),
					"other":     tftypes.NewValue(tftypes.String, "newvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Object-Attribute": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"scratch_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"interface": tftypes.NewValue(tftypes.String, "SCSI"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"scratch_disk": {
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"interface": types.StringType,
								},
							},
							Optional: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface"),
			val:  "NVME",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"scratch_disk": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"scratch_disk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"interface": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"interface": tftypes.NewValue(tftypes.String, "NVME"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Set": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags":  tftypes.Set{ElementType: tftypes.String},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"tags": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"tags": {
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("tags"),
			val:  []string{"one", "two"},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags":  tftypes.Set{ElementType: tftypes.String},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Set-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk1"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: SetNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, SetNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":                   tftypes.String,
					"delete_with_instance": tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "disk1"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
			})),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk0"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Set-Element-duplicate": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags": tftypes.Set{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"tags": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "one"),
						tftypes.NewValue(tftypes.String, "two"),
						tftypes.NewValue(tftypes.String, "three"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"tags": {
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("tags").WithElementKeyValue(tftypes.NewValue(tftypes.String, "three")),
			val:  "three",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags": tftypes.Set{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
					tftypes.NewValue(tftypes.String, "three"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-String": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "originalvalue"),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
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
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			val:  "newvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.String,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.String, "newvalue"),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"write-root": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Bool,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type:     types.BoolType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath(),
			val:  false,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Bool,
				},
			}, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\nexpected tftypes.Object[\"test\":tftypes.Bool], got tftypes.Bool",
				),
			},
		},
		"write-Bool": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.Bool,
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type:     types.BoolType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			val:  false,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.Bool,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.Bool, false),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-List": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags": tftypes.List{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"tags": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("tags"),
			val:  []string{"one", "two"},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags":  tftypes.List{ElementType: tftypes.String},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-List-AttrTypeWithValidateWarning-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: testtypes.ListTypeWithValidateWarning{
								ListType: types.ListType{
									ElemType: types.StringType,
								},
							},
							Optional: true,
							Computed: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0),
			val:  "testvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.List{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test")),
			},
		},
		"write-List-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-List-Element-length-error": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
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
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(1),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("disks"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 2 as list currently has 0 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"write-List-Element-AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: ListNestedAttributes(map[string]Attribute{
								"id": {
									Type:     testtypes.StringTypeWithValidateWarning{},
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
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0).WithAttributeName("id")),
			},
		},
		"write-Map": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			val: map[string]string{
				"newkey": "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"newkey": tftypes.NewValue(tftypes.String, "newvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Map-AttrTypeWithValidateWarning-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: testtypes.MapTypeWithValidateWarning{
								MapType: types.MapType{
									ElemType: types.StringType,
								},
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("key"),
			val:  "keyvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key": tftypes.NewValue(tftypes.String, "keyvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test")),
			},
		},
		"write-Map-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.MapType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("key"),
			val:  "keyvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key": tftypes.NewValue(tftypes.String, "keyvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Map-Element-AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.MapType{
								ElemType: testtypes.StringTypeWithValidateWarning{},
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("key"),
			val:  "keyvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key": tftypes.NewValue(tftypes.String, "keyvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyString("key")),
			},
		},
		"write-Number": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.Number,
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type:     types.NumberType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			val:  1,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.Number,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.Number, 1),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Object": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
							},
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"scratch_disk": {
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"interface": types.StringType,
								},
							},
							Optional: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("scratch_disk"),
			val: struct {
				Interface string `tfsdk:"interface"`
			}{
				Interface: "NVME",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"scratch_disk": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"scratch_disk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"interface": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"interface": tftypes.NewValue(tftypes.String, "NVME"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Set": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags":  tftypes.Set{ElementType: tftypes.String},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"tags": {
							Type: types.SetType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("tags"),
			val:  []string{"one", "two"},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags":  tftypes.Set{ElementType: tftypes.String},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Set-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: SetNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, SetNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":                   tftypes.String,
					"delete_with_instance": tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
			})),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Set-AttrTypeWithValidateWarning-Element": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: testtypes.SetTypeWithValidateWarning{
								SetType: types.SetType{
									ElemType: types.StringType,
								},
							},
							Optional: true,
							Computed: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyValue(tftypes.NewValue(tftypes.String, "testvalue")),
			val:  "testvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Set{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test")),
			},
		},
		"write-Set-Element-AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: SetNestedAttributes(map[string]Attribute{
								"id": {
									Type:     testtypes.StringTypeWithValidateWarning{},
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, SetNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id":                   tftypes.String,
					"delete_with_instance": tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
			})),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyValue(tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"id":                   tftypes.String,
						"delete_with_instance": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
					"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
				})).WithAttributeName("id")),
			},
		},
		"write-String": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.String,
					},
				}, nil),
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
			path: tftypes.NewAttributePath().WithAttributeName("test"),
			val:  "newvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.String,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.String, "newvalue"),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
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
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "originalname"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
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
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newname"),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name")),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.plan.SetAttribute(context.Background(), tc.path, tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				for _, diagnostic := range diags {
					t.Log(diagnostic)
				}
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.plan.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
