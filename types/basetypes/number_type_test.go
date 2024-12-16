// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

func TestNumberTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: NewNumberValue(big.NewFloat(123)),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: NewNumberUnknown(),
		},
		"unknown-with-notnull-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness: tfrefinement.NewNullness(false),
			}),
			expectation: NewNumberUnknown().RefineAsNotNull(),
		},
		"unknown-with-lowerbound-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(1.23), true),
			}),
			expectation: NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
		},
		"unknown-with-upperbound-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(4.56), false),
			}),
			expectation: NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), false),
		},
		"unknown-with-both-bound-refinements": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(1.23), true),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(4.56), false),
			}),
			expectation: NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.56), false),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: NewNumberNull(),
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *big.Float, expected *big.Float",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := NumberType{}.ValueFromTerraform(ctx, test.input)
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

func TestNumberTypeValueFromTerraform_RefinementNullCollapse(t *testing.T) {
	t.Parallel()

	// This shouldn't happen, but this test ensures that if we receive this kind of refinement, that we will
	// convert it to a known null value.
	input := tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
		tfrefinement.KeyNullness: tfrefinement.NewNullness(true),
	})
	expectation := NewNumberNull()

	got, err := NumberType{}.ValueFromTerraform(context.Background(), input)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !got.Equal(expectation) {
		t.Errorf("Expected %+v, got %+v", expectation, got)
	}
}
