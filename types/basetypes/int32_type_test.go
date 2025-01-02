// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"

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
		"unknown-with-notnull-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness: tfrefinement.NewNullness(false),
			}),
			expectation: NewInt32Unknown().RefineAsNotNull(),
		},
		"unknown-with-lowerbound-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(10), true),
			}),
			expectation: NewInt32Unknown().RefineWithLowerBound(10, true),
		},
		"unknown-with-upperbound-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(100), false),
			}),
			expectation: NewInt32Unknown().RefineWithUpperBound(100, false),
		},
		"unknown-with-both-bound-refinements": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(10), true),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(100), false),
			}),
			expectation: NewInt32Unknown().RefineWithLowerBound(10, true).RefineWithUpperBound(100, false),
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
		name, test := name, test
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

func TestInt32TypeValueFromTerraform_RefinementNullCollapse(t *testing.T) {
	t.Parallel()

	// This shouldn't happen, but this test ensures that if we receive this kind of refinement, that we will
	// convert it to a known null value.
	input := tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
		tfrefinement.KeyNullness: tfrefinement.NewNullness(true),
	})
	expectation := NewInt32Null()

	got, err := Int32Type{}.ValueFromTerraform(context.Background(), input)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !got.Equal(expectation) {
		t.Errorf("Expected %+v, got %+v", expectation, got)
	}
}
