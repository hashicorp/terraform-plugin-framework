// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschema_test

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaBlockPathExpressions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema   fwschema.Schema
		expected path.Expressions
	}{
		"no-blocks": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			expected: path.Expressions{},
		},
		"blocks": {
			schema: testschema.Schema{
				Blocks: map[string]fwschema.Block{
					"list_block": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"test_block_attribute": testschema.Attribute{
									Required: true,
									Type:     types.StringType,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeList,
					},
					"set_block": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"test_block_attribute": testschema.Attribute{
									Required: true,
									Type:     types.StringType,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeSet,
					},
					"single_block": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"test_block_attribute": testschema.Attribute{
									Required: true,
									Type:     types.StringType,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeSingle,
					},
				},
			},
			expected: path.Expressions{
				path.MatchRoot("list_block"),
				path.MatchRoot("set_block"),
				path.MatchRoot("single_block"),
			},
		},
		"nested-blocks": {
			schema: testschema.Schema{
				Blocks: map[string]fwschema.Block{
					"list_block": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"test_block_attribute": testschema.Attribute{
									Required: true,
									Type:     types.StringType,
								},
							},
							Blocks: map[string]fwschema.Block{
								"nested_list_block": testschema.Block{
									NestedObject: testschema.NestedBlockObject{
										Attributes: map[string]fwschema.Attribute{
											"test_block_attribute": testschema.Attribute{
												Required: true,
												Type:     types.StringType,
											},
										},
									},
									NestingMode: fwschema.BlockNestingModeList,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeList,
					},
					"set_block": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"test_block_attribute": testschema.Attribute{
									Required: true,
									Type:     types.StringType,
								},
							},
							Blocks: map[string]fwschema.Block{
								"nested_set_block": testschema.Block{
									NestedObject: testschema.NestedBlockObject{
										Attributes: map[string]fwschema.Attribute{
											"test_block_attribute": testschema.Attribute{
												Required: true,
												Type:     types.StringType,
											},
										},
									},
									NestingMode: fwschema.BlockNestingModeSet,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeSet,
					},
					"single_block": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"test_block_attribute": testschema.Attribute{
									Required: true,
									Type:     types.StringType,
								},
							},
							Blocks: map[string]fwschema.Block{
								"nested_single_block": testschema.Block{
									NestedObject: testschema.NestedBlockObject{
										Attributes: map[string]fwschema.Attribute{
											"test_block_attribute": testschema.Attribute{
												Required: true,
												Type:     types.StringType,
											},
										},
									},
									NestingMode: fwschema.BlockNestingModeSingle,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeSingle,
					},
				},
			},
			expected: path.Expressions{
				path.MatchRoot("list_block"),
				path.MatchRoot("list_block").AtAnyListIndex().AtName("nested_list_block"),
				path.MatchRoot("set_block"),
				path.MatchRoot("set_block").AtAnySetValue().AtName("nested_set_block"),
				path.MatchRoot("single_block"),
				path.MatchRoot("single_block").AtName("nested_single_block"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwschema.SchemaBlockPathExpressions(context.Background(), testCase.schema)

			// Prevent differences due to randomized Go map access during testing.
			sort.Slice(testCase.expected, func(i, j int) bool {
				return testCase.expected[i].String() < testCase.expected[j].String()
			})

			sort.Slice(got, func(i, j int) bool {
				return got[i].String() < got[j].String()
			})

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
