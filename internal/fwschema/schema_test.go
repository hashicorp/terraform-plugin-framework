// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwschema_test

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

func TestSchemaAttributeAtTerraformPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        fwschema.Schema
		path          *tftypes.AttributePath
		expected      fwschema.Attribute
		expectedError error
	}{
		"empty": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{},
			},
			path:          tftypes.NewAttributePath(),
			expected:      nil,
			expectedError: errors.New("unexpected type testschema.Schema"),
		},
		"string-attribute-exact": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test_attribute"),
			expected: testschema.Attribute{
				Required: true,
				Type:     types.StringType,
			},
		},
		"string-attribute-ErrInvalidStep": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("test_attribute").WithElementKeyInt(0),
			expected:      nil,
			expectedError: errors.New("ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.StringType"),
		},
		"dynamic-attribute-exact": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.DynamicType,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test_attribute"),
			expected: testschema.Attribute{
				Required: true,
				Type:     types.DynamicType,
			},
		},
		"dynamic-attribute-ErrPathInsideDynamicAttribute": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.DynamicType,
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("test_attribute").WithElementKeyInt(0),
			expected:      nil,
			expectedError: fwschema.ErrPathInsideDynamicAttribute,
		},
		"object-attribute-exact": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test_attribute"),
			expected: testschema.Attribute{
				Required: true,
				Type: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"dynamic": types.DynamicType,
					},
				},
			},
		},
		"object-attribute-dynamic-type-ErrPathInsideAtomicAttribute": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("test_attribute").WithAttributeName("dynamic"),
			expected:      nil,
			expectedError: fwschema.ErrPathInsideAtomicAttribute,
		},
		"object-attribute-dynamic-type-ErrPathInsideDynamicAttribute": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("test_attribute").WithAttributeName("dynamic").WithElementKeyInt(0),
			expected:      nil,
			expectedError: fwschema.ErrPathInsideDynamicAttribute,
		},
		"block-ErrPathIsBlock": {
			schema: testschema.Schema{
				Blocks: map[string]fwschema.Block{
					"test_block": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"test_attribute": testschema.Attribute{
									Optional: true,
									Type:     types.StringType,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeList,
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("test_block"),
			expected:      nil,
			expectedError: fwschema.ErrPathIsBlock,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := fwschema.SchemaAttributeAtTerraformPath(context.Background(), testCase.schema, testCase.path)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("Unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSchemaTypeAtTerraformPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		schema        fwschema.Schema
		path          *tftypes.AttributePath
		expected      attr.Type
		expectedError error
	}{
		"empty": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{},
			},
			path:     tftypes.NewAttributePath(),
			expected: types.ObjectType{},
		},
		"string-attribute-exact": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test_attribute"),
			expected: types.StringType,
		},
		"string-attribute-ErrInvalidStep": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("test_attribute").WithElementKeyInt(0),
			expected:      nil,
			expectedError: errors.New("ElementKeyInt(0) still remains in the path: cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.StringType"),
		},
		"dynamic-attribute-exact": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.DynamicType,
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test_attribute"),
			expected: types.DynamicType,
		},
		"dynamic-attribute-ErrPathInsideDynamicAttribute": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type:     types.DynamicType,
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("test_attribute").WithElementKeyInt(0),
			expected:      nil,
			expectedError: fwschema.ErrPathInsideDynamicAttribute,
		},
		"object-attribute-exact": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test_attribute"),
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"dynamic": types.DynamicType,
				},
			},
		},
		"object-attribute-dynamic-type": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("test_attribute").WithAttributeName("dynamic"),
			expected: types.DynamicType,
		},
		"object-attribute-dynamic-type-ErrPathInsideDynamicAttribute": {
			schema: testschema.Schema{
				Attributes: map[string]fwschema.Attribute{
					"test_attribute": testschema.Attribute{
						Required: true,
						Type: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"dynamic": types.DynamicType,
							},
						},
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("test_attribute").WithAttributeName("dynamic").WithElementKeyInt(0),
			expected:      nil,
			expectedError: fwschema.ErrPathInsideDynamicAttribute,
		},
		"block": {
			schema: testschema.Schema{
				Blocks: map[string]fwschema.Block{
					"test_block": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"test_attribute": testschema.Attribute{
									Optional: true,
									Type:     types.StringType,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeList,
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("test_block"),
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_attribute": types.StringType,
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := fwschema.SchemaTypeAtTerraformPath(context.Background(), testCase.schema, testCase.path)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("Unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}
