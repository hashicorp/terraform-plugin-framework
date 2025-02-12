// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestObjectTypeAttributeTypes_immutable(t *testing.T) {
	t.Parallel()

	typ := ObjectType{AttrTypes: map[string]attr.Type{"test": StringType{}}}
	typ.AttributeTypes()["test"] = BoolType{}

	if !typ.Equal(ObjectType{AttrTypes: map[string]attr.Type{"test": StringType{}}}) {
		t.Fatal("unexpected AttributeTypes mutation")
	}
}

func TestObjectTypeTerraformType_simple(t *testing.T) {
	t.Parallel()
	result := ObjectType{
		AttrTypes: map[string]attr.Type{
			"foo": StringType{},
			"bar": NumberType{},
			"baz": BoolType{},
		},
	}.TerraformType(context.Background())
	if diff := cmp.Diff(result, tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
			"baz": tftypes.Bool,
		},
	}); diff != "" {
		t.Errorf("unexpected result (+expected, -got): %s", diff)
	}
}

func TestObjectTypeTerraformType_empty(t *testing.T) {
	t.Parallel()
	result := ObjectType{}.TerraformType(context.Background())
	if diff := cmp.Diff(result, tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}); diff != "" {
		t.Errorf("unexpected result (+expected, -got): %s", diff)
	}
}

func TestObjectTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    ObjectType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"basic-object": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
					"b": BoolType{},
					"c": NumberType{},
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Bool,
					"c": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "red"),
				"b": tftypes.NewValue(tftypes.Bool, true),
				"c": tftypes.NewValue(tftypes.Number, 123),
			}),
			expected: NewObjectValueMust(
				map[string]attr.Type{
					"a": StringType{},
					"b": BoolType{},
					"c": NumberType{},
				},
				map[string]attr.Value{
					"a": NewStringValue("red"),
					"b": NewBoolValue(true),
					"c": NewNumberValue(big.NewFloat(123)),
				},
			),
		},
		"basic-object-dynamic-types": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": DynamicType{},
					"b": DynamicType{},
					"c": DynamicType{},
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.DynamicPseudoType,
					"b": tftypes.DynamicPseudoType,
					"c": tftypes.DynamicPseudoType,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "red"),
				"b": tftypes.NewValue(tftypes.Bool, true),
				"c": tftypes.NewValue(tftypes.Number, 123),
			}),
			expected: NewObjectValueMust(
				map[string]attr.Type{
					"a": DynamicType{},
					"b": DynamicType{},
					"c": DynamicType{},
				},
				map[string]attr.Value{
					"a": NewDynamicValue(NewStringValue("red")),
					"b": NewDynamicValue(NewBoolValue(true)),
					"c": NewDynamicValue(NewNumberValue(big.NewFloat(123))),
				},
			),
		},
		"extra-attribute": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "red"),
				"b": tftypes.NewValue(tftypes.Bool, true),
			}),
			expectedErr: `expected tftypes.Object["a":tftypes.String], got tftypes.Object["a":tftypes.String, "b":tftypes.Bool]`,
		},
		"missing-attribute": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
					"b": BoolType{},
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "red"),
			}),
			expectedErr: `expected tftypes.Object["a":tftypes.String, "b":tftypes.Bool], got tftypes.Object["a":tftypes.String]`,
		},
		"wrong-type": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
				},
			},
			input:       tftypes.NewValue(tftypes.String, "hello"),
			expectedErr: `expected tftypes.Object["a":tftypes.String], got tftypes.String`,
		},
		"nil-type": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
				},
			},
			input: tftypes.NewValue(nil, nil),
			expected: NewObjectNull(
				map[string]attr.Type{
					"a": StringType{},
				},
			),
		},
		"unknown": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, tftypes.UnknownValue),
			expected: NewObjectUnknown(
				map[string]attr.Type{
					"a": StringType{},
				},
			),
		},
		"null": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, nil),
			expected: NewObjectNull(
				map[string]attr.Type{
					"a": StringType{},
				},
			),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := test.receiver.ValueFromTerraform(context.Background(), test.input)
			if err != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err.Error())
					return
				}
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, err.Error())
					return
				}
			}
			if test.expectedErr != "" && err == nil {
				t.Errorf("Expected err to be %q, got nil", test.expectedErr)
				return
			}
			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("unexpected result (-expected, +got): %s", diff)
			}
			if test.expected != nil && test.expected.IsNull() != test.input.IsNull() {
				t.Errorf("Expected null-ness match: expected %t, got %t", test.expected.IsNull(), test.input.IsNull())
			}
			if test.expected != nil && test.expected.IsUnknown() != !test.input.IsKnown() {
				t.Errorf("Expected unknown-ness match: expected %t, got %t", test.expected.IsUnknown(), !test.input.IsKnown())
			}
		})
	}
}

func TestObjectTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver ObjectType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			expected: true,
		},
		"missing-attr": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			expected: false,
		},
		"extra-attr": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			expected: false,
		},
		"diff-attrs": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"e": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			expected: false,
		},
		"diff": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": BoolType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			expected: false,
		},
		"nested-diff": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: StringType{},
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType{},
				"b": NumberType{},
				"c": BoolType{},
				"d": ListType{
					ElemType: BoolType{},
				},
			}},
			expected: false,
		},
		"wrongType": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
				},
			},
			input:    NumberType{},
			expected: false,
		},
		"nil": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType{},
				},
			},
			input:    nil,
			expected: false,
		},
		"nil-attrs": {
			receiver: ObjectType{},
			input:    ObjectType{},
			expected: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if test.expected != got {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestObjectTypeString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ObjectType
		expected string
	}{
		"AttrTypes-empty": {
			input:    ObjectType{AttrTypes: map[string]attr.Type{}},
			expected: "types.ObjectType[]",
		},
		"AttrTypes-known": {
			input:    ObjectType{AttrTypes: map[string]attr.Type{"testattr": StringType{}}},
			expected: "types.ObjectType[\"testattr\":basetypes.StringType]",
		},
		"AttrTypes-missing": {
			input:    ObjectType{},
			expected: "types.ObjectType[]", // intentionally similar to empty
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
