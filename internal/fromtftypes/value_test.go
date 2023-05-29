// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromtftypes_test

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tfType        tftypes.Value
		attrType      attr.Type
		expected      attr.Value
		expectedError error
	}{
		"empty-tftype": {
			tfType:   tftypes.Value{},
			attrType: types.BoolType,
			expected: types.BoolNull(),
		},
		"nil-attr-type": {
			tfType:        tftypes.Value{},
			attrType:      nil,
			expected:      nil,
			expectedError: fmt.Errorf("unable to convert tftypes.Value (invalid typeless tftypes.Value<>) to attr.Value: missing attr.Type"),
		},
		"invalid-attr-type": {
			tfType:        tftypes.NewValue(tftypes.Bool, true),
			attrType:      testtypes.InvalidType{},
			expected:      nil,
			expectedError: fmt.Errorf("unable to convert tftypes.Value (tftypes.Bool<\"true\">) to attr.Value: intentional ValueFromTerraform error"),
		},
		"bool-null": {
			tfType:   tftypes.NewValue(tftypes.Bool, nil),
			attrType: types.BoolType,
			expected: types.BoolNull(),
		},
		"bool-unknown": {
			tfType:   tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			attrType: types.BoolType,
			expected: types.BoolUnknown(),
		},
		"bool-value": {
			tfType:   tftypes.NewValue(tftypes.Bool, true),
			attrType: types.BoolType,
			expected: types.BoolValue(true),
		},
		"float64-null": {
			tfType:   tftypes.NewValue(tftypes.Number, nil),
			attrType: types.Float64Type,
			expected: types.Float64Null(),
		},
		"float64-unknown": {
			tfType:   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			attrType: types.Float64Type,
			expected: types.Float64Unknown(),
		},
		"float64-value": {
			tfType:   tftypes.NewValue(tftypes.Number, big.NewFloat(1.2)),
			attrType: types.Float64Type,
			expected: types.Float64Value(1.2),
		},
		"int64-null": {
			tfType:   tftypes.NewValue(tftypes.Number, nil),
			attrType: types.Int64Type,
			expected: types.Int64Null(),
		},
		"int64-unknown": {
			tfType:   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			attrType: types.Int64Type,
			expected: types.Int64Unknown(),
		},
		"int64-value": {
			tfType:   tftypes.NewValue(tftypes.Number, 123),
			attrType: types.Int64Type,
			expected: types.Int64Value(123),
		},
		"list-null": {
			tfType: tftypes.NewValue(
				tftypes.List{
					ElementType: tftypes.String,
				},
				nil,
			),
			attrType: types.ListType{
				ElemType: types.StringType,
			},
			expected: types.ListNull(types.StringType),
		},
		"list-unknown": {
			tfType: tftypes.NewValue(
				tftypes.List{
					ElementType: tftypes.String,
				},
				tftypes.UnknownValue,
			),
			attrType: types.ListType{
				ElemType: types.StringType,
			},
			expected: types.ListUnknown(types.StringType),
		},
		"list-value": {
			tfType: tftypes.NewValue(
				tftypes.List{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, "test-value"),
				},
			),
			attrType: types.ListType{
				ElemType: types.StringType,
			},
			expected: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("test-value"),
				},
			),
		},
		"number-null": {
			tfType:   tftypes.NewValue(tftypes.Number, nil),
			attrType: types.NumberType,
			expected: types.NumberNull(),
		},
		"number-unknown": {
			tfType:   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			attrType: types.NumberType,
			expected: types.NumberUnknown(),
		},
		"number-value": {
			tfType:   tftypes.NewValue(tftypes.Number, big.NewFloat(1.2)),
			attrType: types.NumberType,
			expected: types.NumberValue(big.NewFloat(1.2)),
		},
		"object-null": {
			tfType: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attr": tftypes.String,
					},
				},
				nil,
			),
			attrType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
			},
			expected: types.ObjectNull(
				map[string]attr.Type{
					"test_attr": types.StringType,
				},
			),
		},
		"object-unknown": {
			tfType: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attr": tftypes.String,
					},
				},
				tftypes.UnknownValue,
			),
			attrType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
			},
			expected: types.ObjectUnknown(
				map[string]attr.Type{
					"test_attr": types.StringType,
				},
			),
		},
		"object-value": {
			tfType: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attr": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"test_attr": tftypes.NewValue(tftypes.String, "test-value"),
				},
			),
			attrType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
			},
			expected: types.ObjectValueMust(
				map[string]attr.Type{
					"test_attr": types.StringType,
				},
				map[string]attr.Value{
					"test_attr": types.StringValue("test-value"),
				},
			),
		},
		"set-null": {
			tfType: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				nil,
			),
			attrType: types.SetType{
				ElemType: types.StringType,
			},
			expected: types.SetNull(types.StringType),
		},
		"set-unknown": {
			tfType: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				tftypes.UnknownValue,
			),
			attrType: types.SetType{
				ElemType: types.StringType,
			},
			expected: types.SetUnknown(types.StringType),
		},
		"set-value": {
			tfType: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, "test-value"),
				},
			),
			attrType: types.SetType{
				ElemType: types.StringType,
			},
			expected: types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("test-value"),
				},
			),
		},
		"string-null": {
			tfType:   tftypes.NewValue(tftypes.String, nil),
			attrType: types.StringType,
			expected: types.StringNull(),
		},
		"string-unknown": {
			tfType:   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			attrType: types.StringType,
			expected: types.StringUnknown(),
		},
		"string-value": {
			tfType:   tftypes.NewValue(tftypes.String, "test-value"),
			attrType: types.StringType,
			expected: types.StringValue("test-value"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := fromtftypes.Value(context.Background(), testCase.tfType, testCase.attrType)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, tfType: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
