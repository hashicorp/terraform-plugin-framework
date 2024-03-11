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

func TestDynamicTypeEqual(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		receiver DynamicType
		input    attr.Type
		expected bool
	}{
		"equal": {
			receiver: DynamicType{},
			input:    DynamicType{},
			expected: true,
		},
		"wrong-type": {
			receiver: DynamicType{},
			input:    StringType{},
			expected: false,
		},
		"nil": {
			receiver: DynamicType{},
			input:    nil,
			expected: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if test.expected != got {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestDynamicTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}{
		"dynamic-bool-to-dynamic": {
			input:       tftypes.NewValue(tftypes.DynamicPseudoType, true),
			expectedErr: "ambiguous known value for `tftypes.DynamicPseudoType` detected",
		},
		"null-to-dynamic": {
			input:    tftypes.NewValue(tftypes.DynamicPseudoType, nil),
			expected: NewDynamicNull(),
		},
		"unknown-to-dynamic": {
			input:    tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
			expected: NewDynamicUnknown(),
		},
		"bool-to-dynamic": {
			input:    tftypes.NewValue(tftypes.Bool, true),
			expected: NewDynamicValue(NewBoolValue(true)),
		},
		"number-to-dynamic": {
			input:    tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
			expected: NewDynamicValue(NewNumberValue(big.NewFloat(1.2345))),
		},
		"string-to-dynamic": {
			input:    tftypes.NewValue(tftypes.String, "hello world"),
			expected: NewDynamicValue(NewStringValue("hello world")),
		},
		"list-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.List{
					ElementType: tftypes.Bool,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Bool, true),
					tftypes.NewValue(tftypes.Bool, false),
					tftypes.NewValue(tftypes.Bool, true),
				},
			),
			expected: NewDynamicValue(
				NewListValueMust(
					BoolType{},
					[]attr.Value{
						NewBoolValue(true),
						NewBoolValue(false),
						NewBoolValue(true),
					},
				),
			),
		},
		"set-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.Number,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Number, big.NewFloat(1.2345)),
					tftypes.NewValue(tftypes.Number, big.NewFloat(678)),
					tftypes.NewValue(tftypes.Number, big.NewFloat(9.1)),
				},
			),
			expected: NewDynamicValue(
				NewSetValueMust(
					NumberType{},
					[]attr.Value{
						NewNumberValue(big.NewFloat(1.2345)),
						NewNumberValue(big.NewFloat(678)),
						NewNumberValue(big.NewFloat(9.1)),
					},
				),
			),
		},
		"map-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key1": tftypes.NewValue(tftypes.String, "hello"),
					"key2": tftypes.NewValue(tftypes.String, "world"),
				},
			),
			expected: NewDynamicValue(
				NewMapValueMust(
					StringType{},
					map[string]attr.Value{
						"key1": NewStringValue("hello"),
						"key2": NewStringValue("world"),
					},
				),
			),
		},
		"object-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"attr1": tftypes.Bool,
						"attr2": tftypes.String,
						"attr3": tftypes.Number,
						"attr4": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested_attr1": tftypes.String,
							},
						},
					},
				}, map[string]tftypes.Value{
					"attr1": tftypes.NewValue(tftypes.Bool, true),
					"attr2": tftypes.NewValue(tftypes.String, "hello"),
					"attr3": tftypes.NewValue(tftypes.Number, big.NewFloat(9.1)),
					"attr4": tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested_attr1": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"nested_attr1": tftypes.NewValue(tftypes.String, "world"),
						}),
				},
			),
			expected: NewDynamicValue(
				NewObjectValueMust(
					map[string]attr.Type{
						"attr1": BoolType{},
						"attr2": StringType{},
						"attr3": NumberType{},
						"attr4": ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr1": StringType{},
							},
						},
					},
					map[string]attr.Value{
						"attr1": NewBoolValue(true),
						"attr2": NewStringValue("hello"),
						"attr3": NewNumberValue(big.NewFloat(9.1)),
						"attr4": NewObjectValueMust(
							map[string]attr.Type{
								"nested_attr1": StringType{},
							},
							map[string]attr.Value{
								"nested_attr1": NewStringValue("world"),
							},
						),
					},
				),
			),
		},
		"object-with-dpt-null-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"attr1": tftypes.Bool,
						"attr2": tftypes.String,
						"attr3": tftypes.Number,
						"attr4": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"attr1": tftypes.NewValue(tftypes.Bool, nil),
					"attr2": tftypes.NewValue(tftypes.String, "hello"),
					"attr3": tftypes.NewValue(tftypes.Number, big.NewFloat(9.1)),
					"attr4": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
				},
			),
			expected: NewDynamicValue(
				NewObjectValueMust(
					map[string]attr.Type{
						"attr1": BoolType{},
						"attr2": StringType{},
						"attr3": NumberType{},
						"attr4": DynamicType{},
					},
					map[string]attr.Value{
						"attr1": NewBoolNull(),
						"attr2": NewStringValue("hello"),
						"attr3": NewNumberValue(big.NewFloat(9.1)),
						"attr4": NewDynamicNull(),
					},
				),
			),
		},
		"object-with-dpt-unknown-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"attr1": tftypes.Bool,
						"attr2": tftypes.String,
						"attr3": tftypes.Number,
						"attr4": tftypes.DynamicPseudoType,
					},
				}, map[string]tftypes.Value{
					"attr1": tftypes.NewValue(tftypes.Bool, nil),
					"attr2": tftypes.NewValue(tftypes.String, "hello"),
					"attr3": tftypes.NewValue(tftypes.Number, big.NewFloat(9.1)),
					"attr4": tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
				},
			),
			expected: NewDynamicValue(
				NewObjectValueMust(
					map[string]attr.Type{
						"attr1": BoolType{},
						"attr2": StringType{},
						"attr3": NumberType{},
						"attr4": DynamicType{},
					},
					map[string]attr.Value{
						"attr1": NewBoolNull(),
						"attr2": NewStringValue("hello"),
						"attr3": NewNumberValue(big.NewFloat(9.1)),
						"attr4": NewDynamicUnknown(),
					},
				),
			),
		},
		"tuple-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.Tuple{
					ElementTypes: []tftypes.Type{
						tftypes.Bool,
						tftypes.String,
						tftypes.Number,
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested_attr1": tftypes.String,
							},
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Bool, nil),
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.Number, big.NewFloat(9.1)),
					tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested_attr1": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"nested_attr1": tftypes.NewValue(tftypes.String, "world"),
						},
					),
				},
			),
			expected: NewDynamicValue(
				NewTupleValueMust(
					[]attr.Type{
						BoolType{},
						StringType{},
						NumberType{},
						ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr1": StringType{},
							},
						},
					},
					[]attr.Value{
						NewBoolNull(),
						NewStringValue("hello"),
						NewNumberValue(big.NewFloat(9.1)),
						NewObjectValueMust(
							map[string]attr.Type{
								"nested_attr1": StringType{},
							},
							map[string]attr.Value{
								"nested_attr1": NewStringValue("world"),
							},
						),
					},
				),
			),
		},
		"tuple-with-dpt-null-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.Tuple{
					ElementTypes: []tftypes.Type{
						tftypes.Bool,
						tftypes.String,
						tftypes.Number,
						tftypes.DynamicPseudoType,
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Bool, nil),
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.Number, big.NewFloat(9.1)),
					tftypes.NewValue(tftypes.DynamicPseudoType, nil),
				},
			),
			expected: NewDynamicValue(
				NewTupleValueMust(
					[]attr.Type{
						BoolType{},
						StringType{},
						NumberType{},
						DynamicType{},
					},
					[]attr.Value{
						NewBoolNull(),
						NewStringValue("hello"),
						NewNumberValue(big.NewFloat(9.1)),
						NewDynamicNull(),
					},
				),
			),
		},
		"tuple-with-dpt-unknown-to-dynamic": {
			input: tftypes.NewValue(
				tftypes.Tuple{
					ElementTypes: []tftypes.Type{
						tftypes.Bool,
						tftypes.String,
						tftypes.Number,
						tftypes.DynamicPseudoType,
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Bool, nil),
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.Number, big.NewFloat(9.1)),
					tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
				},
			),
			expected: NewDynamicValue(
				NewTupleValueMust(
					[]attr.Type{
						BoolType{},
						StringType{},
						NumberType{},
						DynamicType{},
					},
					[]attr.Value{
						NewBoolNull(),
						NewStringValue("hello"),
						NewNumberValue(big.NewFloat(9.1)),
						NewDynamicUnknown(),
					},
				),
			),
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := DynamicType{}.ValueFromTerraform(context.Background(), test.input)
			if gotErr != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", gotErr.Error())
					return
				}
				if gotErr.Error() != test.expectedErr {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, gotErr.Error())
					return
				}
			}
			if gotErr == nil && test.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", test.expectedErr)
				return
			}
			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("Unexpected diff (-expected, +got): %s", diff)
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
