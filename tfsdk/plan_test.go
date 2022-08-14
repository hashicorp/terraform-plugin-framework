package tfsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	intreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPlanGet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		plan          Plan
		target        any
		expected      any
		expectedDiags diag.Diagnostics
	}{
		// Refer to fwschemadata.TestDataGet for more exhaustive unit testing.
		// These test cases are to ensure Plan schema and data values are
		// passed appropriately to the shared implementation.
		"valid": {
			plan: Plan{
				Raw: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"string": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"string": tftypes.NewValue(tftypes.String, "test"),
					},
				),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"string": {
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
			},
			target: new(struct {
				String types.String `tfsdk:"string"`
			}),
			expected: &struct {
				String types.String `tfsdk:"string"`
			}{
				String: types.String{Value: "test"},
			},
		},
		"diagnostic": {
			plan: Plan{
				Raw: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool": tftypes.NewValue(tftypes.Bool, nil),
					},
				),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"bool": {
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
			},
			target: new(struct {
				String types.String `tfsdk:"bool"`
			}),
			expected: &struct {
				String types.String `tfsdk:"bool"`
			}{
				String: types.String{},
			},
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Root("bool"),
					intreflect.DiagNewAttributeValueIntoWrongType{
						ValType:    reflect.TypeOf(types.Bool{}),
						TargetType: reflect.TypeOf(types.String{}),
						SchemaType: types.BoolType,
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.plan.Get(context.Background(), testCase.target)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(testCase.target, testCase.expected); diff != "" {
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
		// Refer to fwschemadata.TestDataGetAtPath for more exhaustive unit
		// testing. These test cases are to ensure Plan schema and data values
		// are passed appropriately to the shared implementation.
		"valid": {
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
		"diagnostics": {
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
			expected:      &testtypes.String{InternalString: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(path.Root("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.plan.GetAttribute(context.Background(), path.Root("name"), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected, cmp.Transformer("testtypes", func(in *testtypes.String) testtypes.String { return *in }), cmp.Transformer("types", func(in *types.String) types.String { return *in })); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanPathMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		plan          Plan
		expression    path.Expression
		expected      path.Paths
		expectedDiags diag.Diagnostics
	}{
		// Refer to fwschemadata.TestDataPathMatches for more exhaustive unit testing.
		// These test cases are to ensure Plan schema and data values are
		// passed appropriately to the shared implementation.
		"AttributeNameExact-match": {
			plan: Plan{
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.StringType,
						},
					},
				},
				Raw: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "test-value"),
					},
				),
			},
			expression: path.MatchRoot("test"),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-mismatch": {
			plan: Plan{
				Schema: Schema{
					Attributes: map[string]Attribute{
						"test": {
							Type: types.StringType,
						},
					},
				},
				Raw: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "test-value"),
					},
				),
			},
			expression: path.MatchRoot("not-test"),
			expected:   nil,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: not-test",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := testCase.plan.PathMatches(context.Background(), testCase.expression)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
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
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(path.Root("name"))},
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
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(path.Root("name"))},
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
		path          path.Path
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
							}),
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
			path: path.Root("disks").AtListIndex(1),
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
							}),
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
			path: path.Root("disks").AtListIndex(2),
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
					path.Root("disks"),
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
							}),
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
			path: path.Root("disks").AtListIndex(0),
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
							}),
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
			path: path.Root("disks").AtListIndex(1),
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
					path.Root("disks"),
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
			path: path.Root("test").AtMapKey("key2"),
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
			path: path.Root("test").AtMapKey("key"),
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
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "mynewdisk"},
					"delete_with_instance": types.Bool{Value: true},
				},
			}),
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
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "mynewdisk"},
					"delete_with_instance": types.Bool{Value: true},
				},
			}),
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
			path: path.Root("test"),
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
			path: path.Root("tags"),
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
							}),
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
			path: path.Root("disks").AtListIndex(1),
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
			path: path.Root("test"),
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
			path: path.Root("test").AtMapKey("key"),
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
			path: path.Root("test"),
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
			path: path.Root("scratch_disk"),
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
			path: path.Root("scratch_disk").AtName("interface"),
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
			path: path.Root("tags"),
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
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "disk1"},
					"delete_with_instance": types.Bool{Value: false},
				},
			}),
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
			path: path.Root("tags").AtSetValue(types.String{Value: "three"}),
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
			path: path.Root("test"),
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
			path: path.Empty(),
			val:  false,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Bool,
				},
			}, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
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
			path: path.Root("test"),
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
			path: path.Root("tags"),
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
			path: path.Root("test").AtListIndex(0),
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
				testtypes.TestWarningDiagnostic(path.Root("test")),
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
							}),
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
			path: path.Root("disks").AtListIndex(0),
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
							}),
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
			path: path.Root("disks").AtListIndex(1),
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
					path.Root("disks"),
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
							}),
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
			path: path.Root("disks").AtListIndex(0),
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
				testtypes.TestWarningDiagnostic(path.Root("disks").AtListIndex(0).AtName("id")),
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
			path: path.Root("test"),
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
			path: path.Root("test").AtMapKey("key"),
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
				testtypes.TestWarningDiagnostic(path.Root("test")),
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
			path: path.Root("test").AtMapKey("key"),
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
			path: path.Root("test").AtMapKey("key"),
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
				testtypes.TestWarningDiagnostic(path.Root("test").AtMapKey("key")),
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
			path: path.Root("test"),
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
			path: path.Root("scratch_disk"),
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
			path: path.Root("tags"),
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
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "mynewdisk"},
					"delete_with_instance": types.Bool{Value: true},
				},
			}),
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
			path: path.Root("test").AtSetValue(types.String{Value: "testvalue"}),
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
				testtypes.TestWarningDiagnostic(path.Root("test")),
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
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "mynewdisk"},
					"delete_with_instance": types.Bool{Value: true},
				},
			}),
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
				testtypes.TestWarningDiagnostic(path.Root("disks").AtSetValue(types.Object{
					AttrTypes: map[string]attr.Type{
						"id":                   types.StringType,
						"delete_with_instance": types.BoolType,
					},
					Attrs: map[string]attr.Value{
						"id":                   types.String{Value: "mynewdisk"},
						"delete_with_instance": types.Bool{Value: true},
					},
				}).AtName("id")),
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
			path: path.Root("test"),
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
			path: path.Root("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "originalname"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(path.Root("name"))},
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
			path: path.Root("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newname"),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("name")),
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
