package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPathMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        Schema
		tfTypeValue   tftypes.Value
		expression    path.Expression
		expected      path.Paths
		expectedDiags diag.Diagnostics
	}{
		"ohno": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.StringType,
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "test-value"),
				},
			),
			expression: path.MatchRoot("test").AtParent().AtParent(),
			expected:   nil,
		},
		"AttributeNameExact-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.StringType,
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "test-value"),
				},
			),
			expression: path.MatchRoot("test"),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-mismatch": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.StringType,
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "test-value"),
				},
			),
			expression: path.MatchRoot("not-test"),
			expected:   nil,
		},
		"AttributeNameExact-ElementKeyIntAny-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtAnyListIndex(),
			expected: path.Paths{
				path.Root("test").AtListIndex(0),
				path.Root("test").AtListIndex(1),
				path.Root("test").AtListIndex(2),
			},
		},
		"AttributeNameExact-ElementKeyIntAny-mismatch": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtAnyListIndex(),
			expected:   nil,
		},
		"AttributeNameExact-ElementKeyIntExact-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtListIndex(1),
			expected: path.Paths{
				path.Root("test").AtListIndex(1),
			},
		},
		"AttributeNameExact-ElementKeyIntExact-AttributeNameExact-Parent-AttributeNameExact-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test_parent": {
						Attributes: ListNestedAttributes(map[string]Attribute{
							"test_child1": {
								Type: types.StringType,
							},
							"test_child2": {
								Type: types.StringType,
							},
						}),
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_parent": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_child1": tftypes.String,
									"test_child2": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_parent": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_child1": tftypes.String,
									"test_child2": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_child1": tftypes.String,
										"test_child2": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test_child1": tftypes.NewValue(tftypes.String, "test-value-list-0-child-1"),
									"test_child2": tftypes.NewValue(tftypes.String, "test-value-list-0-child-2"),
								},
							),
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_child1": tftypes.String,
										"test_child2": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test_child1": tftypes.NewValue(tftypes.String, "test-value-list-1-child-1"),
									"test_child2": tftypes.NewValue(tftypes.String, "test-value-list-1-child-2"),
								},
							),
						},
					),
				},
			),
			// e.g. Something that would be created in an attribute plan modifier or validator
			expression: path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1").AtParent().AtName("test_child2"),
			expected: path.Paths{
				path.Root("test_parent").AtListIndex(1).AtName("test_child2"),
			},
		},
		"AttributeNameExact-ElementKeyIntExact-mismatch": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtListIndex(4),
			expected:   nil,
		},
		"AttributeNameExact-ElementKeyStringAny-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							// Map access is non-deterministic, so test with
							// a single key to prevent ordering issues in
							// the expected path.Paths
							"test-key1": tftypes.NewValue(tftypes.String, "test-value1"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtAnyMapKey(),
			expected: path.Paths{
				path.Root("test").AtMapKey("test-key1"),
			},
		},
		"AttributeNameExact-ElementKeyStringAny-mismatch": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtAnyMapKey(),
			expected:   nil,
		},
		"AttributeNameExact-ElementKeyStringExact-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"test-key1": tftypes.NewValue(tftypes.String, "test-value1"),
							"test-key2": tftypes.NewValue(tftypes.String, "test-value2"),
							"test-key3": tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtMapKey("test-key2"),
			expected: path.Paths{
				path.Root("test").AtMapKey("test-key2"),
			},
		},
		"AttributeNameExact-ElementKeyStringExact-mismatch": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.Map{
							ElementType: tftypes.String,
						},
						map[string]tftypes.Value{
							"test-key1": tftypes.NewValue(tftypes.String, "test-value1"),
							"test-key2": tftypes.NewValue(tftypes.String, "test-value2"),
							"test-key3": tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtMapKey("test-key4"),
			expected:   nil,
		},
		"AttributeNameExact-ElementKeyValueAny-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtAnySetValue(),
			expected: path.Paths{
				path.Root("test").AtSetValue(types.String{Value: "test-value1"}),
				path.Root("test").AtSetValue(types.String{Value: "test-value2"}),
				path.Root("test").AtSetValue(types.String{Value: "test-value3"}),
			},
		},
		"AttributeNameExact-ElementKeyValueAny-mismatch": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtAnySetValue(),
			expected:   nil,
		},
		"AttributeNameExact-ElementKeyValueExact-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtSetValue(types.String{Value: "test-value2"}),
			expected: path.Paths{
				path.Root("test").AtSetValue(types.String{Value: "test-value2"}),
			},
		},
		"AttributeNameExact-ElementKeyValueExact-mismatch": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test": {
						Type: types.SetType{
							ElemType: types.StringType,
						},
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.String,
						},
						[]tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-value1"),
							tftypes.NewValue(tftypes.String, "test-value2"),
							tftypes.NewValue(tftypes.String, "test-value3"),
						},
					),
				},
			),
			expression: path.MatchRoot("test").AtSetValue(types.String{Value: "test-value4"}),
			expected:   nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := pathMatches(context.Background(), testCase.schema, testCase.tfTypeValue, testCase.expression)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
