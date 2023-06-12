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

func TestMapTypeElementType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    MapType
		expected attr.Type
	}{
		"ElemType-known": {
			input:    MapType{ElemType: StringType{}},
			expected: StringType{},
		},
		"ElemType-missing": {
			input:    MapType{},
			expected: missingType{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ElementType()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMapTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    MapType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"map-of-strings": {
			input: MapType{
				ElemType: StringType{},
			},
			expected: tftypes.Map{
				ElementType: tftypes.String,
			},
		},
		"map-of-map-of-strings": {
			input: MapType{
				ElemType: MapType{
					ElemType: StringType{},
				},
			},
			expected: tftypes.Map{
				ElementType: tftypes.Map{
					ElementType: tftypes.String,
				},
			},
		},
		"map-of-map-of-map-of-strings": {
			input: MapType{
				ElemType: MapType{
					ElemType: MapType{
						ElemType: StringType{},
					},
				},
			},
			expected: tftypes.Map{
				ElementType: tftypes.Map{
					ElementType: tftypes.Map{
						ElementType: tftypes.String,
					},
				},
			},
		},
		"ElemType-missing": {
			input: MapType{},
			expected: tftypes.Map{
				ElementType: tftypes.DynamicPseudoType,
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.TerraformType(context.Background())
			if !got.Equal(test.expected) {
				t.Errorf("Expected %s, got %s", test.expected, got)
			}
		})
	}
}

func TestMapTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    MapType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"basic-map": {
			receiver: MapType{
				ElemType: NumberType{},
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, map[string]tftypes.Value{
				"one":   tftypes.NewValue(tftypes.Number, 1),
				"two":   tftypes.NewValue(tftypes.Number, 2),
				"three": tftypes.NewValue(tftypes.Number, 3),
			}),
			expected: NewMapValueMust(
				NumberType{},
				map[string]attr.Value{
					"one":   NewNumberValue(big.NewFloat(1)),
					"two":   NewNumberValue(big.NewFloat(2)),
					"three": NewNumberValue(big.NewFloat(3)),
				},
			),
		},
		"wrong-type": {
			receiver: MapType{
				ElemType: NumberType{},
			},
			input:       tftypes.NewValue(tftypes.String, "wrong"),
			expectedErr: `can't use tftypes.String<"wrong"> as value of MapValue, can only use tftypes.Map values`,
		},
		"nil-type": {
			receiver: MapType{
				ElemType: NumberType{},
			},
			input:    tftypes.NewValue(nil, nil),
			expected: NewMapNull(NumberType{}),
		},
		"missing-element-type": {
			receiver: MapType{},
			input: tftypes.NewValue(
				tftypes.Map{
					ElementType: tftypes.String,
				},
				map[string]tftypes.Value{
					"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
				},
			),
			expectedErr: `can't use tftypes.Map[tftypes.String]<"testkey":tftypes.String<"testvalue">> as value of Map with ElementType basetypes.missingType, can only use tftypes.DynamicPseudoType values`,
		},
		"unknown": {
			receiver: MapType{
				ElemType: NumberType{},
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, tftypes.UnknownValue),
			expected: NewMapUnknown(NumberType{}),
		},
		"null": {
			receiver: MapType{
				ElemType: NumberType{},
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, nil),
			expected: NewMapNull(NumberType{}),
		},
	}

	for name, test := range tests {
		name, test := name, test
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

func TestMapTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver MapType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: MapType{
				ElemType: ListType{
					ElemType: StringType{},
				},
			},
			input: MapType{
				ElemType: ListType{
					ElemType: StringType{},
				},
			},
			expected: true,
		},
		"diff": {
			receiver: MapType{
				ElemType: ListType{
					ElemType: StringType{},
				},
			},
			input: MapType{
				ElemType: ListType{
					ElemType: NumberType{},
				},
			},
			expected: false,
		},
		"wrongType": {
			receiver: MapType{
				ElemType: StringType{},
			},
			input:    NumberType{},
			expected: false,
		},
		"nil": {
			receiver: MapType{
				ElemType: StringType{},
			},
			input:    nil,
			expected: false,
		},
		"nil-elem": {
			receiver: MapType{},
			input:    MapType{},
			// MapTypes with nil ElemTypes are invalid, and aren't
			// equal to anything
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

func TestMapTypeString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    MapType
		expected string
	}{
		"ElemType-known": {
			input:    MapType{ElemType: StringType{}},
			expected: "types.MapType[basetypes.StringType]",
		},
		"ElemType-missing": {
			input:    MapType{},
			expected: "types.MapType[!!! MISSING TYPE !!!]",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
