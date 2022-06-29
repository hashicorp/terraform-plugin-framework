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
			expected: types.Bool{Null: true},
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
			expected: types.Bool{Null: true},
		},
		"bool-unknown": {
			tfType:   tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			attrType: types.BoolType,
			expected: types.Bool{Unknown: true},
		},
		"bool-value": {
			tfType:   tftypes.NewValue(tftypes.Bool, true),
			attrType: types.BoolType,
			expected: types.Bool{Value: true},
		},
		"float64-null": {
			tfType:   tftypes.NewValue(tftypes.Number, nil),
			attrType: types.Float64Type,
			expected: types.Float64{Null: true},
		},
		"float64-unknown": {
			tfType:   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			attrType: types.Float64Type,
			expected: types.Float64{Unknown: true},
		},
		"float64-value": {
			tfType:   tftypes.NewValue(tftypes.Number, big.NewFloat(1.2)),
			attrType: types.Float64Type,
			expected: types.Float64{Value: 1.2},
		},
		"int64-null": {
			tfType:   tftypes.NewValue(tftypes.Number, nil),
			attrType: types.Int64Type,
			expected: types.Int64{Null: true},
		},
		"int64-unknown": {
			tfType:   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			attrType: types.Int64Type,
			expected: types.Int64{Unknown: true},
		},
		"int64-value": {
			tfType:   tftypes.NewValue(tftypes.Number, 123),
			attrType: types.Int64Type,
			expected: types.Int64{Value: 123},
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
			expected: types.List{
				ElemType: types.StringType,
				Null:     true,
			},
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
			expected: types.List{
				ElemType: types.StringType,
				Unknown:  true,
			},
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
			expected: types.List{
				ElemType: types.StringType,
				Elems: []attr.Value{
					types.String{Value: "test-value"},
				},
			},
		},
		"number-null": {
			tfType:   tftypes.NewValue(tftypes.Number, nil),
			attrType: types.NumberType,
			expected: types.Number{Null: true},
		},
		"number-unknown": {
			tfType:   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			attrType: types.NumberType,
			expected: types.Number{Unknown: true},
		},
		"number-value": {
			tfType:   tftypes.NewValue(tftypes.Number, big.NewFloat(1.2)),
			attrType: types.NumberType,
			expected: types.Number{Value: big.NewFloat(1.2)},
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
			expected: types.Object{
				AttrTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
				Null: true,
			},
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
			expected: types.Object{
				AttrTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
				Unknown: true,
			},
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
			expected: types.Object{
				AttrTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"test_attr": types.String{Value: "test-value"},
				},
			},
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
			expected: types.Set{
				ElemType: types.StringType,
				Null:     true,
			},
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
			expected: types.Set{
				ElemType: types.StringType,
				Unknown:  true,
			},
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
			expected: types.Set{
				ElemType: types.StringType,
				Elems: []attr.Value{
					types.String{Value: "test-value"},
				},
			},
		},
		"string-null": {
			tfType:   tftypes.NewValue(tftypes.String, nil),
			attrType: types.StringType,
			expected: types.String{Null: true},
		},
		"string-unknown": {
			tfType:   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			attrType: types.StringType,
			expected: types.String{Unknown: true},
		},
		"string-value": {
			tfType:   tftypes.NewValue(tftypes.String, "test-value"),
			attrType: types.StringType,
			expected: types.String{Value: "test-value"},
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
