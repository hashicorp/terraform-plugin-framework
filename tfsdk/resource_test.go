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
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestResourceGet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		resource      tfsdk.Resource
		target        any
		expected      any
		expectedDiags diag.Diagnostics
	}{
		// Refer to fwschemadata.TestDataGet for more exhaustive unit testing.
		// These test cases are to ensure Resource schema and data values are
		// passed appropriately to the shared implementation.
		"valid": {
			resource: tfsdk.Resource{
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
							RequiredForImport: true,
							Type:              types.StringType,
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
			resource: tfsdk.Resource{
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
							RequiredForImport: true,
							Type:              types.BoolType,
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.resource.Get(context.Background(), testCase.target)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(testCase.target, testCase.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestResourceGetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		resource      tfsdk.Resource
		target        interface{}
		expected      interface{}
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		// Refer to fwschemadata.TestDataGetAtPath for more exhaustive unit
		// testing. These test cases are to ensure Resource schema and data values
		// are passed appropriately to the shared implementation.
		"valid": {
			resource: tfsdk.Resource{
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
							Type:              types.StringType,
							RequiredForImport: true,
						},
					},
				},
			},
			target:   new(string),
			expected: pointer("namevalue"),
		},
		"diagnostics": {
			resource: tfsdk.Resource{
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
							Type:              testtypes.StringTypeWithValidateWarning{},
							RequiredForImport: true,
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.resource.GetAttribute(context.Background(), path.Root("name"), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected, cmp.Transformer("testtypes", func(in *testtypes.String) testtypes.String { return *in }), cmp.Transformer("types", func(in *types.String) types.String { return *in })); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestResourcePathMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		resource      tfsdk.Resource
		expression    path.Expression
		expected      path.Paths
		expectedDiags diag.Diagnostics
	}{
		// Refer to fwschemadata.TestDataPathMatches for more exhaustive unit testing.
		// These test cases are to ensure Resource schema and data values are
		// passed appropriately to the shared implementation.
		"AttributeNameExact-match": {
			resource: tfsdk.Resource{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:              types.StringType,
							RequiredForImport: true,
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
			resource: tfsdk.Resource{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Type:              types.StringType,
							RequiredForImport: true,
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
					"Invalid Path Expression for Schema",
					"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: not-test",
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := testCase.resource.PathMatches(context.Background(), testCase.expression)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestResourceSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		resource      tfsdk.Resource
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		// Refer to fwschemadata.TestDataSet for more exhaustive unit testing.
		// These test cases are to ensure Resource schema and data values are
		// passed appropriately to the shared implementation.
		"valid": {
			resource: tfsdk.Resource{
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
							Type:              types.StringType,
							RequiredForImport: true,
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
			resource: tfsdk.Resource{
				Raw: tftypes.Value{},
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"name": testschema.Attribute{
							Type:              testtypes.StringTypeWithValidateWarning{},
							RequiredForImport: true,
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.resource.Set(context.Background(), tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.resource.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestResourceSetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		resource      tfsdk.Resource
		path          path.Path
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		// Refer to fwschemadata.TestDataSetAtPath for more exhaustive unit
		// testing. These test cases are to ensure ResourceIdentity schema and data values
		// are passed appropriately to the shared implementation.
		"valid": {
			resource: tfsdk.Resource{
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
							Type:              types.StringType,
							RequiredForImport: true,
						},
						"other": testschema.Attribute{
							Type:              types.StringType,
							OptionalForImport: true,
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
			resource: tfsdk.Resource{
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
							Type:              testtypes.StringTypeWithValidateWarning{},
							RequiredForImport: true,
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.resource.SetAttribute(context.Background(), tc.path, tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				for _, diagnostic := range diags {
					t.Log(diagnostic)
				}
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.resource.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
