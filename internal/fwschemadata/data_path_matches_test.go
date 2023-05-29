// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDataPathMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        fwschema.Schema
		tfTypeValue   tftypes.Value
		expression    path.Expression
		expected      path.Paths
		expectedDiags diag.Diagnostics
	}{
		"AttributeNameExact-match": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
					"Invalid Path Expression for Schema",
					"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: not-test",
				),
			},
		},
		"AttributeNameExact-AttributeNameExact-match": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_parent": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"test_child": testschema.Attribute{
									Type: types.StringType,
								},
							},
						},
						NestingMode: fwschema.NestingModeSingle,
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_parent": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"test_child": testschema.Attribute{
									Type: types.StringType,
								},
							},
						},
						NestingMode: fwschema.NestingModeSingle,
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
					"Invalid Path Expression for Schema",
					"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test_parent.not_test_child",
				),
			},
		},
		"AttributeNameExact-AttributeNameExact-parent-null": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_parent": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"test_child": testschema.Attribute{
									Type: types.StringType,
								},
							},
						},
						NestingMode: fwschema.NestingModeSingle,
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_parent": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"test_child": testschema.Attribute{
									Type: types.StringType,
								},
							},
						},
						NestingMode: fwschema.NestingModeSingle,
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
					"Invalid Path Expression for Schema",
					"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[*]",
				),
			},
		},
		"AttributeNameExact-ElementKeyIntAny-parent-null": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_parent": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"test_child1": testschema.Attribute{
									Type: types.StringType,
								},
								"test_child2": testschema.Attribute{
									Type: types.StringType,
								},
							},
						},
						NestingMode: fwschema.NestingModeList,
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_parent": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"test_child1": testschema.Attribute{
									Type: types.StringType,
								},
								"test_child2": testschema.Attribute{
									Type: types.StringType,
								},
							},
						},
						NestingMode: fwschema.NestingModeList,
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_parent": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"test_child1": testschema.Attribute{
									Type: types.StringType,
								},
								"test_child2": testschema.Attribute{
									Type: types.StringType,
								},
							},
						},
						NestingMode: fwschema.NestingModeList,
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
		"AttributeNameExact-ElementKeyIntExact-parent-null": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
					"Invalid Path Expression for Schema",
					"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[\"*\"]",
				),
			},
		},
		"AttributeNameExact-ElementKeyStringAny-parent-null": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
		"AttributeNameExact-ElementKeyStringExact-parent-null": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
				path.Root("test").AtSetValue(types.StringValue("test-value1")),
				path.Root("test").AtSetValue(types.StringValue("test-value2")),
				path.Root("test").AtSetValue(types.StringValue("test-value3")),
			},
		},
		"AttributeNameExact-ElementKeyValueAny-mismatch": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
					"Invalid Path Expression for Schema",
					"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test[Value(*)]",
				),
			},
		},
		"AttributeNameExact-ElementKeyValueAny-parent-null": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value2")),
			expected: path.Paths{
				path.Root("test").AtSetValue(types.StringValue("test-value2")),
			},
		},
		"AttributeNameExact-ElementKeyValueExact-mismatch": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value4")),
			expected:   nil,
		},
		"AttributeNameExact-ElementKeyValueExact-parent-null": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value2")),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-ElementKeyValueExact-parent-unknown": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value2")),
			expected: path.Paths{
				path.Root("test"),
			},
		},
		"AttributeNameExact-Parent": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
					"Invalid Path Expression for Schema",
					"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
						"This can happen if the path expression does not correctly follow the schema in structure or types. "+
						"Please report this to the provider developers.\n\n"+
						"Path Expression: test.<",
				),
			},
		},
		"AttributeNameExact-Parent-Parent": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test": testschema.Attribute{
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
					"Invalid Path Expression for Schema",
					"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
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

			data := fwschemadata.Data{
				Schema:         testCase.schema,
				TerraformValue: testCase.tfTypeValue,
			}

			got, diags := data.PathMatches(context.Background(), testCase.expression)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
