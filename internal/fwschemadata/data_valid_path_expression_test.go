// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataValidPathExpression(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data       fwschemadata.Data
		expression path.Expression
		expected   bool
	}{
		"resolved-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.StringType,
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtParent().AtName("test"),
			expected:   true,
		},
		"AttributeNameExact-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.StringType,
						},
					},
				},
			},
			expression: path.MatchRoot("test"),
			expected:   true,
		},
		"AttributeNameExact-mismatch": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.StringType,
						},
					},
				},
			},
			expression: path.MatchRoot("not_test"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"nested_test": types.StringType,
								},
							},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtName("nested_test"),
			expected:   true,
		},
		"AttributeNameExact-AtListIndexAny-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtAnyListIndex(),
			expected:   true,
		},
		"AttributeNameExact-AtListIndexAny-mismatch-type": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.SetType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtAnyListIndex(),
			expected:   false,
		},
		"AttributeNameExact-AtListIndexExact-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtListIndex(1),
			expected:   true,
		},
		"AttributeNameExact-AtListIndexExact-mismatch-type": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.SetType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtListIndex(1),
			expected:   false,
		},
		"AttributeNameExact-AtMapKeyAny-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.MapType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtAnyMapKey(),
			expected:   true,
		},
		"AttributeNameExact-AtMapKeyAny-mismatch-type": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtAnyMapKey(),
			expected:   false,
		},
		"AttributeNameExact-AtMapKeyExact-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.MapType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtMapKey("test-key"),
			expected:   true,
		},
		"AttributeNameExact-AtMapKeyExact-mismatch-type": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtMapKey("test-key"),
			expected:   false,
		},
		"AttributeNameExact-AtSetValueAny-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.SetType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtAnySetValue(),
			expected:   true,
		},
		"AttributeNameExact-AtSetValueAny-mismatch-type": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtAnySetValue(),
			expected:   false,
		},
		"AttributeNameExact-AtSetValueExact-match": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.SetType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value")),
			expected:   true,
		},
		"AttributeNameExact-AtSetValueExact-mismatch-type": {
			data: fwschemadata.Data{
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Required: true,
							Type:     types.ListType{ElemType: types.StringType},
						},
					},
				},
			},
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value")),
			expected:   false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.data.ValidPathExpression(context.Background(), testCase.expression)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
