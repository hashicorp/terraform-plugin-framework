// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func TestFloat32TypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	var v float32 = 123.456

	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value-int": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: NewFloat32Value(float32(123.0)),
		},
		"value-float": {
			input:       tftypes.NewValue(tftypes.Number, float64(v)),
			expectation: NewFloat32Value(v),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: NewFloat32Unknown(),
		},
		"unknown-with-notnull-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness: tfrefinement.NewNullness(false),
			}),
			expectation: NewFloat32Unknown().RefineAsNotNull(),
		},
		"unknown-with-lowerbound-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(1.23), true),
			}),
			expectation: NewFloat32Unknown().RefineWithLowerBound(1.23, true),
		},
		"unknown-with-upperbound-refinement": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(4.56), false),
			}),
			expectation: NewFloat32Unknown().RefineWithUpperBound(4.56, false),
		},
		"unknown-with-both-bound-refinements": {
			input: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(1.23), true),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(4.56), false),
			}),
			expectation: NewFloat32Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.56, false),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: NewFloat32Null(),
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *big.Float, expected *big.Float",
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/327
		// To ensure underlying *big.Float precision matches, create `expectation` via struct literal
		"zero-string-float": {
			input: tftypes.NewValue(tftypes.Number, testMustParseFloat("0.0")),
			expectation: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.0"),
			},
		},
		"positive-string-float": {
			input: tftypes.NewValue(tftypes.Number, testMustParseFloat("123.2")),
			expectation: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("123.2"),
			},
		},
		"negative-string-float": {
			input: tftypes.NewValue(tftypes.Number, testMustParseFloat("-123.2")),
			expectation: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("-123.2"),
			},
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/815
		// To ensure underlying *big.Float precision matches, create `expectation` via struct literal
		"retain-string-float-512-precision": {
			input: tftypes.NewValue(tftypes.Number, testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003")),
			expectation: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003"),
			},
		},
		// Reference: https://pkg.go.dev/math/big#Float.Float32
		// Reference: https://pkg.go.dev/math#pkg-constants
		"SmallestNonzeroFloat32": {
			input:       tftypes.NewValue(tftypes.Number, big.NewFloat(math.SmallestNonzeroFloat32)),
			expectation: NewFloat32Value(math.SmallestNonzeroFloat32),
		},
		"SmallestNonzeroFloat32-below": {
			input:       tftypes.NewValue(tftypes.Number, testMustParseFloat("1.401298464324817070923729583289916131280e-46")),
			expectedErr: fmt.Sprintf("Value %s cannot be represented as a 32-bit floating point.", testMustParseFloat("1.401298464324817070923729583289916131280e-46")),
		},
		// Reference: https://pkg.go.dev/math/big#Float.Float32
		// Reference: https://pkg.go.dev/math#pkg-constants
		"MaxFloat32": {
			input:       tftypes.NewValue(tftypes.Number, big.NewFloat(math.MaxFloat32)),
			expectation: NewFloat32Value(math.MaxFloat32),
		},
		"MaxFloat32-above": {
			input:       tftypes.NewValue(tftypes.Number, testMustParseFloat("3.40282346638528859811704183484516925440e+39")),
			expectedErr: fmt.Sprintf("Value %s cannot be represented as a 32-bit floating point.", testMustParseFloat("3.40282346638528859811704183484516925440e+39")),
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := Float32Type{}.ValueFromTerraform(ctx, test.input)
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
			if diff := cmp.Diff(got, test.expectation); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
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

func TestFloat32TypeValueFromTerraform_RefinementNullCollapse(t *testing.T) {
	t.Parallel()

	// This shouldn't happen, but this test ensures that if we receive this kind of refinement, that we will
	// convert it to a known null value.
	input := tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
		tfrefinement.KeyNullness: tfrefinement.NewNullness(true),
	})
	expectation := NewFloat32Null()

	got, err := Float32Type{}.ValueFromTerraform(context.Background(), input)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !got.Equal(expectation) {
		t.Errorf("Expected %+v, got %+v", expectation, got)
	}
}
