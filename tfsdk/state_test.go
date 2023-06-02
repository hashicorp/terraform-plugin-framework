// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfsdk_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	intreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStateGet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		state         tfsdk.State
		target        any
		expected      any
		expectedDiags diag.Diagnostics
	}{
		// Refer to fwschemadata.TestDataGet for more exhaustive unit testing.
		// These test cases are to ensure Plan schema and data values are
		// passed appropriately to the shared implementation.
		"valid": {
			state: tfsdk.State{
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
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"string": testschema.Attribute{
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
				String: types.StringValue("test"),
			},
		},
		"diagnostic": {
			state: tfsdk.State{
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
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool": testschema.Attribute{
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

			diags := testCase.state.Get(context.Background(), testCase.target)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(testCase.target, testCase.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStateGetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state         tfsdk.State
		target        interface{}
		expected      interface{}
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		// Refer to fwschemadata.TestDataGetAtPath for more exhaustive unit
		// testing. These test cases are to ensure Plan schema and data values
		// are passed appropriately to the shared implementation.
		"valid": {
			state: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(string),
			expected: pointer("namevalue"),
		},
		"diagnostics": {
			state: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			target:        new(testtypes.String),
			expected:      &testtypes.String{InternalString: types.StringValue("namevalue"), CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(path.Root("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.state.GetAttribute(context.Background(), path.Root("name"), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected, cmp.Transformer("testtypes", func(in *testtypes.String) testtypes.String { return *in }), cmp.Transformer("types", func(in *types.String) types.String { return *in })); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStateSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state         tfsdk.State
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		// Refer to fwschemadata.TestDataSet for more exhaustive unit testing.
		// These test cases are to ensure State schema and data values are
		// passed appropriately to the shared implementation.
		"valid": {
			state: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "oldvalue"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
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
		"diagnostics": {
			state: tfsdk.State{
				Raw: tftypes.Value{},
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
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

			diags := tc.state.Set(context.Background(), tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.state.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStateSetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state         tfsdk.State
		path          path.Path
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		// Refer to fwschemadata.TestDataSetAtPath for more exhaustive unit
		// testing. These test cases are to ensure State schema and data values
		// are passed appropriately to the shared implementation.
		"valid": {
			state: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "originalvalue"),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:     types.StringType,
							Required: true,
						},
						"other": testschema.Attribute{
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
		"diagnostics": {
			state: tfsdk.State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
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

			diags := tc.state.SetAttribute(context.Background(), tc.path, tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.state.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
