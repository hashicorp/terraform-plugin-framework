package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPathMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        fwschema.Schema
		tfTypeValue   tftypes.Value
		expression    path.Expression
		expected      path.Paths
		expectedDiags diag.Diagnostics
	}{
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
		"AttributeNameExact-AttributeNameExact-match": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test_parent": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"test_child": {
								Type: types.StringType,
							},
						}),
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_parent": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_child": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_parent": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_child": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"test_child": tftypes.NewValue(tftypes.String, "test-value"),
						},
					),
				},
			),
			expression: path.MatchRoot("test_parent").AtName("test_child"),
			expected: path.Paths{
				path.Root("test_parent").AtName("test_child"),
			},
		},
		"AttributeNameExact-AttributeNameExact-mismatch": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test_parent": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"test_child": {
								Type: types.StringType,
							},
						}),
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_parent": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_child": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_parent": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_child": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"test_child": tftypes.NewValue(tftypes.String, "test-value"),
						},
					),
				},
			),
			expression: path.MatchRoot("test_parent").AtName("not_test_child"),
			expected:   nil,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test_parent.not_test_child",
				),
			},
		},
		"AttributeNameExact-AttributeNameExact-parent-null": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test_parent": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"test_child": {
								Type: types.StringType,
							},
						}),
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_parent": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_child": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_parent": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_child": tftypes.String,
							},
						},
						nil,
					),
				},
			),
			expression: path.MatchRoot("test_parent").AtName("test_child"),
			expected: path.Paths{
				path.Root("test_parent"),
			},
		},
		"AttributeNameExact-AttributeNameExact-parent-unknown": {
			schema: Schema{
				Attributes: map[string]Attribute{
					"test_parent": {
						Attributes: SingleNestedAttributes(map[string]Attribute{
							"test_child": {
								Type: types.StringType,
							},
						}),
					},
				},
			},
			tfTypeValue: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_parent": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_child": tftypes.String,
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_parent": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test_child": tftypes.String,
							},
						},
						tftypes.UnknownValue,
					),
				},
			),
			expression: path.MatchRoot("test_parent").AtName("test_child"),
			expected: path.Paths{
				path.Root("test_parent"),
			},
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
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[*]",
				),
			},
		},
		"AttributeNameExact-ElementKeyIntAny-parent-null": {
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
						nil,
					),
				},
			),
			expression: path.MatchRoot("test").AtAnyListIndex(),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-ElementKeyIntAny-parent-unknown": {
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
						tftypes.UnknownValue,
					),
				},
			),
			expression: path.MatchRoot("test").AtAnyListIndex(),
			expected: path.Paths{
				path.Root("test"),
			},
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
		"AttributeNameExact-ElementKeyIntExact-AttributeNameExact-Parent-AttributeNameExact-parent-null": {
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
								nil,
							),
						},
					),
				},
			),
			// e.g. Something that would be created in an attribute plan modifier or validator
			expression: path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1").AtParent().AtName("test_child2"),
			expected: path.Paths{
				path.Root("test_parent").AtListIndex(1),
			},
		},
		"AttributeNameExact-ElementKeyIntExact-AttributeNameExact-Parent-AttributeNameExact-parent-unknown": {
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
								tftypes.UnknownValue,
							),
						},
					),
				},
			),
			// e.g. Something that would be created in an attribute plan modifier or validator
			expression: path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1").AtParent().AtName("test_child2"),
			expected: path.Paths{
				path.Root("test_parent").AtListIndex(1),
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
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[4]",
				),
			},
		},
		"AttributeNameExact-ElementKeyIntExact-parent-null": {
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
						nil,
					),
				},
			),
			expression: path.MatchRoot("test").AtListIndex(1),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-ElementKeyIntExact-parent-unknown": {
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
						tftypes.UnknownValue,
					),
				},
			),
			expression: path.MatchRoot("test").AtListIndex(1),
			expected: path.Paths{
				path.Root("test"),
			},
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
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[\"*\"]",
				),
			},
		},
		"AttributeNameExact-ElementKeyStringAny-parent-null": {
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
						nil,
					),
				},
			),
			expression: path.MatchRoot("test").AtAnyMapKey(),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-ElementKeyStringAny-parent-unknown": {
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
						tftypes.UnknownValue,
					),
				},
			),
			expression: path.MatchRoot("test").AtAnyMapKey(),
			expected: path.Paths{
				path.Root("test"),
			},
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
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[\"test-key4\"]",
				),
			},
		},
		"AttributeNameExact-ElementKeyStringExact-parent-null": {
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
						nil,
					),
				},
			),
			expression: path.MatchRoot("test").AtMapKey("test-key2"),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-ElementKeyStringExact-parent-unknown": {
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
						tftypes.UnknownValue,
					),
				},
			),
			expression: path.MatchRoot("test").AtMapKey("test-key2"),
			expected: path.Paths{
				path.Root("test"),
			},
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
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[Value(*)]",
				),
			},
		},
		"AttributeNameExact-ElementKeyValueAny-parent-null": {
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
						nil,
					),
				},
			),
			expression: path.MatchRoot("test").AtAnySetValue(),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-ElementKeyValueAny-parent-unknown": {
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
						tftypes.UnknownValue,
					),
				},
			),
			expression: path.MatchRoot("test").AtAnySetValue(),
			expected: path.Paths{
				path.Root("test"),
			},
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
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[Value(\"test-value4\")]",
				),
			},
		},
		"AttributeNameExact-ElementKeyValueExact-parent-null": {
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
						nil,
					),
				},
			),
			expression: path.MatchRoot("test").AtSetValue(types.String{Value: "test-value2"}),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-ElementKeyValueExact-parent-unknown": {
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
						tftypes.UnknownValue,
					),
				},
			),
			expression: path.MatchRoot("test").AtSetValue(types.String{Value: "test-value2"}),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-Parent": {
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
			expression: path.MatchRoot("test").AtParent(),
			expected:   nil,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test.<",
				),
			},
		},
		"AttributeNameExact-Parent-Parent": {
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
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Path Expression for Schema Data",
					"The Terraform Provider unexpectedly matched no paths with the given path expression and current schema data. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test.<.<",
				),
			},
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
