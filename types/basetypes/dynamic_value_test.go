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

func TestDynamicValueToTerraformValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    DynamicValue
		expected tftypes.Value
	}{
		"known-primitive": {
			input:    NewDynamicValue(NewStringValue("test")),
			expected: tftypes.NewValue(tftypes.DynamicPseudoType, "test"),
		},
		"known-collection": {
			input: NewDynamicValue(
				NewListValueMust(NumberType{}, []attr.Value{
					NewNumberValue(big.NewFloat(1.234)),
					NewNumberValue(big.NewFloat(100)),
					NewNumberValue(big.NewFloat(222.1)),
				}),
			),
			expected: tftypes.NewValue(
				tftypes.DynamicPseudoType,
				[]tftypes.Value{
					tftypes.NewValue(tftypes.Number, big.NewFloat(1.234)),
					tftypes.NewValue(tftypes.Number, big.NewFloat(100)),
					tftypes.NewValue(tftypes.Number, big.NewFloat(222.1)),
				},
			),
		},
		"known-structural": {
			input: NewDynamicValue(
				NewObjectValueMust(
					map[string]attr.Type{
						"string_val": StringType{},
						"map_val":    MapType{ElemType: NumberType{}},
					},
					map[string]attr.Value{
						"string_val": NewStringValue("hello world"),
						"map_val": NewMapValueMust(
							NumberType{},
							map[string]attr.Value{
								"num1": NewNumberValue(big.NewFloat(1.234)),
								"num2": NewNumberValue(big.NewFloat(100)),
							},
						),
					},
				),
			),
			expected: tftypes.NewValue(
				tftypes.DynamicPseudoType,
				map[string]tftypes.Value{
					"string_val": tftypes.NewValue(tftypes.String, "hello world"),
					"map_val": tftypes.NewValue(
						tftypes.Map{ElementType: tftypes.Number},
						map[string]tftypes.Value{
							"num1": tftypes.NewValue(tftypes.Number, big.NewFloat(1.234)),
							"num2": tftypes.NewValue(tftypes.Number, big.NewFloat(100)),
						},
					),
				},
			),
		},
		"null": {
			input:    NewDynamicNull(),
			expected: tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		},
		"unknown": {
			input:    NewDynamicUnknown(),
			expected: tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := test.input.ToTerraformValue(ctx)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestDynamicValueEqual(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input       DynamicValue
		candidate   attr.Value
		expectation bool
	}{
		"known-known-same-primitive": {
			input:       NewDynamicValue(NewStringValue("hello")),
			candidate:   NewDynamicValue(NewStringValue("hello")),
			expectation: true,
		},
		"known-known-diff-primitive": {
			input:       NewDynamicValue(NewStringValue("hello")),
			candidate:   NewDynamicValue(NewStringValue("goodbye")),
			expectation: false,
		},
		"known-known-same-collection": {
			input: NewDynamicValue(
				NewSetValueMust(NumberType{}, []attr.Value{
					NewNumberValue(big.NewFloat(1.234)),
					NewNumberValue(big.NewFloat(100)),
					NewNumberValue(big.NewFloat(222.1)),
				}),
			),
			candidate: NewDynamicValue(
				NewSetValueMust(NumberType{}, []attr.Value{
					NewNumberValue(big.NewFloat(1.234)),
					NewNumberValue(big.NewFloat(100)),
					NewNumberValue(big.NewFloat(222.1)),
				}),
			),
			expectation: true,
		},
		"known-known-diff-collection": {
			input: NewDynamicValue(
				NewSetValueMust(NumberType{}, []attr.Value{
					NewNumberValue(big.NewFloat(1.234)),
					NewNumberValue(big.NewFloat(100)),
					NewNumberValue(big.NewFloat(222.1)),
				}),
			),
			candidate: NewDynamicValue(
				NewSetValueMust(NumberType{}, []attr.Value{
					NewNumberValue(big.NewFloat(1.234)),
					NewNumberValue(big.NewFloat(23)),
					NewNumberValue(big.NewFloat(222.1)),
				}),
			),
			expectation: false,
		},
		"known-known-same-structural": {
			input: NewDynamicValue(
				NewObjectValueMust(
					map[string]attr.Type{
						"string_val": StringType{},
						"map_val":    MapType{ElemType: NumberType{}},
					},
					map[string]attr.Value{
						"string_val": NewStringValue("hello world"),
						"map_val": NewMapValueMust(
							NumberType{},
							map[string]attr.Value{
								"num1": NewNumberValue(big.NewFloat(1.234)),
								"num2": NewNumberValue(big.NewFloat(100)),
							},
						),
					},
				),
			),
			candidate: NewDynamicValue(
				NewObjectValueMust(
					map[string]attr.Type{
						"string_val": StringType{},
						"map_val":    MapType{ElemType: NumberType{}},
					},
					map[string]attr.Value{
						"string_val": NewStringValue("hello world"),
						"map_val": NewMapValueMust(
							NumberType{},
							map[string]attr.Value{
								"num1": NewNumberValue(big.NewFloat(1.234)),
								"num2": NewNumberValue(big.NewFloat(100)),
							},
						),
					},
				),
			),
			expectation: true,
		},
		"known-known-diff-structural": {
			input: NewDynamicValue(
				NewObjectValueMust(
					map[string]attr.Type{
						"string_val": StringType{},
						"map_val":    MapType{ElemType: NumberType{}},
					},
					map[string]attr.Value{
						"string_val": NewStringValue("hello world"),
						"map_val": NewMapValueMust(
							NumberType{},
							map[string]attr.Value{
								"num1": NewNumberValue(big.NewFloat(1.234)),
								"num2": NewNumberValue(big.NewFloat(100)),
							},
						),
					},
				),
			),
			candidate: NewDynamicValue(
				NewObjectValueMust(
					map[string]attr.Type{
						"string_val": StringType{},
						"map_val":    MapType{ElemType: NumberType{}},
					},
					map[string]attr.Value{
						"string_val": NewStringValue("goodbye!"),
						"map_val": NewMapValueMust(
							NumberType{},
							map[string]attr.Value{
								"num1": NewNumberValue(big.NewFloat(1.234)),
								"num2": NewNumberValue(big.NewFloat(100)),
							},
						),
					},
				),
			),
			expectation: false,
		},
		"known-unknown": {
			input:       NewDynamicValue(NewStringValue("hello")),
			candidate:   NewDynamicUnknown(),
			expectation: false,
		},
		"known-null": {
			input:       NewDynamicValue(NewStringValue("hello")),
			candidate:   NewDynamicNull(),
			expectation: false,
		},
		"unknown-value": {
			input:       NewDynamicUnknown(),
			candidate:   NewDynamicValue(NewStringValue("hello")),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NewDynamicUnknown(),
			candidate:   NewDynamicUnknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       NewDynamicUnknown(),
			candidate:   NewDynamicNull(),
			expectation: false,
		},
		"null-known": {
			input:       NewDynamicNull(),
			candidate:   NewDynamicValue(NewStringValue("hello")),
			expectation: false,
		},
		"null-unknown": {
			input:       NewDynamicNull(),
			candidate:   NewDynamicUnknown(),
			expectation: false,
		},
		"null-null": {
			input:       NewDynamicNull(),
			candidate:   NewDynamicNull(),
			expectation: true,
		},
		"known-known-no-dynamic-wrapper": {
			input:       NewDynamicValue(NewStringValue("hello")),
			candidate:   NewStringValue("hello"),
			expectation: false,
		},
		"unknown-unknown-no-dynamic-wrapper": {
			input:       NewDynamicUnknown(),
			candidate:   NewStringUnknown(),
			expectation: false,
		},
		"null-null-no-dynamic-wrapper": {
			input:       NewDynamicNull(),
			candidate:   NewStringNull(),
			expectation: false,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.Equal(test.candidate)
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %v, got %v", test.expectation, got)
			}
		})
	}
}
