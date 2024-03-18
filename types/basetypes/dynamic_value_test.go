// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDynamicValueToTerraformValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input         DynamicValue
		expected      tftypes.Value
		expectedError error
	}{
		"known-primitive": {
			input:    NewDynamicValue(NewStringValue("test")),
			expected: tftypes.NewValue(tftypes.String, "test"),
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
				tftypes.List{
					ElementType: tftypes.Number,
				},
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
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"string_val": tftypes.String,
						"map_val": tftypes.Map{
							ElementType: tftypes.Number,
						},
					},
				},
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
		"known-nil-underlying-value": {
			input: DynamicValue{
				value: nil, // Should not panic
				state: attr.ValueStateKnown,
			},
			expected:      tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
			expectedError: errors.New("invalid Dynamic state in ToTerraformValue: DynamicValue is known but the underlying value is unset"),
		},
		"null": {
			input:    NewDynamicNull(),
			expected: tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		},
		"unknown": {
			input:    NewDynamicUnknown(),
			expected: tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
		},
		// For dynamic values, it's possible the underlying type is known but the underlying value itself is null. In this
		// situation, the type information must be preserved when returned back to Terraform.
		"null-value-known-type": {
			input:    NewDynamicValue(NewBoolNull()),
			expected: tftypes.NewValue(tftypes.Bool, nil),
		},
		// For dynamic values, it's possible the underlying type is known but the underlying value itself is unknown. In this
		// situation, the type information must be preserved when returned back to Terraform.
		"unknown-value-known-type": {
			input:    NewDynamicValue(NewListUnknown(StringType{})),
			expected: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := test.input.ToTerraformValue(ctx)
			if err != nil {
				if test.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), test.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", test.expectedError, err)
				}
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
		"known-underlying-value-unset-input": {
			input: DynamicValue{
				value: nil, // Should not panic
				state: attr.ValueStateKnown,
			},
			candidate:   NewDynamicNull(),
			expectation: false,
		},
		"known-underlying-value-unset-candidate": {
			input: NewDynamicNull(),
			candidate: DynamicValue{
				value: nil, // Should not panic
				state: attr.ValueStateKnown,
			},
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
func TestDynamicValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       DynamicValue
		expectation string
	}
	tests := map[string]testCase{
		"known-primitive": {
			input:       NewDynamicValue(NewStringValue("hello world")),
			expectation: `"hello world"`,
		},
		"known-collection": {
			input: NewDynamicValue(NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			)),
			expectation: `["hello","world"]`,
		},
		"known-tuple": {
			input: NewDynamicValue(NewTupleValueMust(
				[]attr.Type{
					StringType{},
					StringType{},
				},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			)),
			expectation: `["hello","world"]`,
		},
		"known-structural": {
			input: NewDynamicValue(NewObjectValueMust(
				map[string]attr.Type{
					"alpha": StringType{},
					"beta":  StringType{},
				},
				map[string]attr.Value{
					"alpha": NewStringValue("hello"),
					"beta":  NewStringValue("world"),
				},
			)),
			expectation: `{"alpha":"hello","beta":"world"}`,
		},
		"known-nil-underlying-value": {
			input: DynamicValue{
				value: nil, // Should not panic
				state: attr.ValueStateKnown,
			},
			expectation: "<unset>",
		},
		"unknown": {
			input:       NewDynamicUnknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewDynamicNull(),
			expectation: "<null>",
		},
		"zero-value": {
			input:       DynamicValue{},
			expectation: "<null>",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.String()
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %q, got %q", test.expectation, got)
			}
		})
	}
}
