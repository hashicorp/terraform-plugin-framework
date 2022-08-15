package tfsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	intreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestConfigGet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config        Config
		target        any
		expected      any
		expectedDiags diag.Diagnostics
	}{
		// Refer to fwschemadata.TestDataGet for more exhaustive unit testing.
		// These test cases are to ensure Plan schema and data values are
		// passed appropriately to the shared implementation.
		"valid": {
			config: Config{
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
			config: Config{
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

			diags := testCase.config.Get(context.Background(), testCase.target)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(testCase.target, testCase.expected); diff != "" {
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
		// Refer to fwschemadata.TestDataGetAtPath for more exhaustive unit
		// testing. These test cases are to ensure Plan schema and data values
		// are passed appropriately to the shared implementation.
		"valid": {
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
		"diagnostics": {
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
			expected:      &testtypes.String{InternalString: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(path.Root("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.config.GetAttribute(context.Background(), path.Root("name"), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected, cmp.Transformer("testtypes", func(in *testtypes.String) testtypes.String { return *in }), cmp.Transformer("types", func(in *types.String) types.String { return *in })); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestConfigPathMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config        Config
		expression    path.Expression
		expected      path.Paths
		expectedDiags diag.Diagnostics
	}{
		// Refer to TestPathMatches for more exhaustive unit testing.
		// These test cases are to ensure Config schema and data values are
		// passed appropriately to the shared implementation.
		"AttributeNameExact-match": {
			config: Config{
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
			config: Config{
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

			got, diags := testCase.config.PathMatches(context.Background(), testCase.expression)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
