// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"math"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func TestInt32TypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: NewInt32Value(123),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: NewInt32Unknown(),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: NewInt32Null(),
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *big.Float, expected *big.Float",
		},
		"Int32Min value": {
			input:       tftypes.NewValue(tftypes.Number, math.MinInt32),
			expectation: NewInt32Value(math.MinInt32),
		},
		"Int32Max value": {
			input:       tftypes.NewValue(tftypes.Number, math.MaxInt32),
			expectation: NewInt32Value(math.MaxInt32),
		},
		"Int32Min - 1: error": {
			input:       tftypes.NewValue(tftypes.Number, math.MinInt32-1),
			expectedErr: "Value %!s(*big.Float=-2147483649) cannot be represented as a 32-bit integer.",
		},
		"Int32Max + 1: error": {
			input:       tftypes.NewValue(tftypes.Number, math.MaxInt32+1),
			expectedErr: "Value %!s(*big.Float=2147483648) cannot be represented as a 32-bit integer.",
		},
		"float value: error": {
			input:       tftypes.NewValue(tftypes.Number, 32.1),
			expectedErr: "Value %!s(*big.Float=32.1) is not an integer.",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := Int32Type{}.ValueFromTerraform(ctx, test.input)
			if err != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if test.expectedErr != err.Error() {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, err.Error())
					return
				}
				// we have an error, and it matches our
				// expectations, we're good
				return
			}
			if test.expectedErr != "" {
				t.Errorf("Expected error to be %q, didn't get an error", test.expectedErr)
				return
			}
			if !got.Equal(test.expectation) {
				t.Errorf("Expected %+v, got %+v", test.expectation, got)
			}
			if test.expectation.IsNull() != test.input.IsNull() {
				t.Errorf("Expected null-ness match: expected %t, got %t", test.expectation.IsNull(), test.input.IsNull())
			}
			if test.expectation.IsUnknown() != !test.input.IsKnown() {
				t.Errorf("Expected unknown-ness match: expected %t, got %t", test.expectation.IsUnknown(), !test.input.IsKnown())
			}
		})
	}
}
