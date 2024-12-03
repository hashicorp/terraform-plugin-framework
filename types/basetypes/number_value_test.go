// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/refinement"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

func numberComparer(i, j *big.Float) bool {
	return (i == nil && j == nil) || (i != nil && j != nil && i.Cmp(j) == 0)
}

func TestNumberValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       NumberValue
		expectation tftypes.Value
	}
	tests := map[string]testCase{
		"value": {
			input:       NewNumberValue(big.NewFloat(123)),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
		},
		"known-nil": {
			input:       NewNumberValue(nil),
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"unknown": {
			input:       NewNumberUnknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"unknown-with-notnull-refinement": {
			input: NewNumberUnknown().RefineAsNotNull(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness: tfrefinement.NewNullness(false),
			}),
		},
		"unknown-with-lower-bound-refinement": {
			input: NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(1.23), true),
			}),
		},
		"unknown-with-upper-bound-refinement": {
			input: NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), false),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(4.56), false),
			}),
		},
		"unknown-with-both-bound-refinements": {
			input: NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.56), false),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(1.23), true),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(4.56), false),
			}),
		},
		"null": {
			input:       NewNumberNull(),
			expectation: tftypes.NewValue(tftypes.Number, nil),
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
			if !cmp.Equal(got, test.expectation, cmp.Comparer(numberComparer)) {
				t.Errorf("Expected %+v, got %+v", test.expectation, got)
			}
		})
	}
}

func TestNumberValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       NumberValue
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       NewNumberValue(big.NewFloat(123)),
			candidate:   NewNumberValue(big.NewFloat(123)),
			expectation: true,
		},
		"known-known-diff": {
			input:       NewNumberValue(big.NewFloat(123)),
			candidate:   NewNumberValue(big.NewFloat(456)),
			expectation: false,
		},
		"known-nil-known": {
			input:       NewNumberValue(nil),
			candidate:   NewNumberValue(big.NewFloat(456)),
			expectation: false,
		},
		"known-nil-null": {
			input:       NewNumberValue(nil),
			candidate:   NewNumberNull(),
			expectation: true,
		},
		"known-unknown": {
			input:       NewNumberValue(big.NewFloat(123)),
			candidate:   NewNumberUnknown(),
			expectation: false,
		},
		"known-null": {
			input:       NewNumberValue(big.NewFloat(123)),
			candidate:   NewNumberNull(),
			expectation: false,
		},
		"known-wrong-type": {
			input:       NewNumberValue(big.NewFloat(123)),
			candidate:   NewFloat64Value(123),
			expectation: false,
		},
		"known-nil": {
			input:       NewNumberValue(big.NewFloat(123)),
			candidate:   nil,
			expectation: false,
		},
		"unknown-known": {
			input:       NewNumberUnknown(),
			candidate:   NewNumberValue(big.NewFloat(123)),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NewNumberUnknown(),
			candidate:   NewNumberUnknown(),
			expectation: true,
		},
		"unknown-unknown-with-notnull-refinement": {
			input:       NewNumberUnknown(),
			candidate:   NewNumberUnknown().RefineAsNotNull(),
			expectation: false,
		},
		"unknown-unknown-with-lowerbound-refinement": {
			input:       NewNumberUnknown(),
			candidate:   NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
			expectation: false,
		},
		"unknown-unknown-with-upperbound-refinement": {
			input:       NewNumberUnknown(),
			candidate:   NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), false),
			expectation: false,
		},
		"unknowns-with-matching-notnull-refinements": {
			input:       NewNumberUnknown().RefineAsNotNull(),
			candidate:   NewNumberUnknown().RefineAsNotNull(),
			expectation: true,
		},
		"unknowns-with-matching-lowerbound-refinements": {
			input:       NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
			candidate:   NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
			expectation: true,
		},
		"unknowns-with-different-lowerbound-refinements": {
			input:       NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
			candidate:   NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.24), true),
			expectation: false,
		},
		"unknowns-with-different-lowerbound-refinements-inclusive": {
			input:       NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
			candidate:   NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), false),
			expectation: false,
		},
		"unknowns-with-matching-upperbound-refinements": {
			input:       NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), true),
			candidate:   NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), true),
			expectation: true,
		},
		"unknowns-with-different-upperbound-refinements": {
			input:       NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), true),
			candidate:   NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.57), true),
			expectation: false,
		},
		"unknowns-with-different-upperbound-refinements-inclusive": {
			input:       NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), true),
			candidate:   NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), false),
			expectation: false,
		},
		"unknowns-with-matching-both-bound-refinements": {
			input:       NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.56), true),
			candidate:   NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.56), true),
			expectation: true,
		},
		"unknowns-with-different-both-bound-refinements": {
			input:       NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.56), true),
			candidate:   NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.57), true),
			expectation: false,
		},
		"unknowns-with-different-both-bound-refinements-inclusive": {
			input:       NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.56), true),
			candidate:   NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.56), false),
			expectation: false,
		},
		"unknown-null": {
			input:       NewNumberUnknown(),
			candidate:   NewNumberNull(),
			expectation: false,
		},
		"unknown-wrong-type": {
			input:       NewNumberUnknown(),
			candidate:   NewFloat64Unknown(),
			expectation: false,
		},
		"unknown-nil": {
			input:       NewNumberUnknown(),
			candidate:   nil,
			expectation: false,
		},
		"null-known": {
			input:       NewNumberNull(),
			candidate:   NewNumberValue(big.NewFloat(123)),
			expectation: false,
		},
		"null-known-nil": {
			input:       NewNumberNull(),
			candidate:   NewNumberValue(nil),
			expectation: true,
		},
		"null-unknown": {
			input:       NewNumberNull(),
			candidate:   NewNumberUnknown(),
			expectation: false,
		},
		"null-null": {
			input:       NewNumberNull(),
			candidate:   NewNumberNull(),
			expectation: true,
		},
		"null-wrong-type": {
			input:       NewNumberNull(),
			candidate:   NewFloat64Null(),
			expectation: false,
		},
		"null-nil": {
			input:       NewNumberNull(),
			candidate:   nil,
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

func TestNumberValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    NumberValue
		expected bool
	}{
		"known": {
			input:    NewNumberValue(big.NewFloat(2.4)),
			expected: false,
		},
		"null": {
			input:    NewNumberNull(),
			expected: true,
		},
		"unknown": {
			input:    NewNumberUnknown(),
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsNull()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNumberValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    NumberValue
		expected bool
	}{
		"known": {
			input:    NewNumberValue(big.NewFloat(2.4)),
			expected: false,
		},
		"null": {
			input:    NewNumberNull(),
			expected: false,
		},
		"unknown": {
			input:    NewNumberUnknown(),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsUnknown()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNumberValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       NumberValue
		expectation string
	}
	tests := map[string]testCase{
		"known-less-than-one": {
			input:       NewNumberValue(big.NewFloat(0.12340984302980000)),
			expectation: "0.123409843",
		},
		"known-more-than-one": {
			input:       NewNumberValue(big.NewFloat(92387938173219.327663)),
			expectation: "9.238793817e+13",
		},
		"known-negative-more-than-one": {
			input:       NewNumberValue(big.NewFloat(-0.12340984302980000)),
			expectation: "-0.123409843",
		},
		"known-negative-less-than-one": {
			input:       NewNumberValue(big.NewFloat(-92387938173219.327663)),
			expectation: "-9.238793817e+13",
		},
		"known-min-float64": {
			input:       NewNumberValue(big.NewFloat(math.SmallestNonzeroFloat64)),
			expectation: "4.940656458e-324",
		},
		"known-max-float64": {
			input:       NewNumberValue(big.NewFloat(math.MaxFloat64)),
			expectation: "1.797693135e+308",
		},
		"unknown": {
			input:       NewNumberUnknown(),
			expectation: "<unknown>",
		},
		"unknown-with-notnull-refinement": {
			input:       NewNumberUnknown().RefineAsNotNull(),
			expectation: "<unknown, not null>",
		},
		"unknown-with-lowerbound-refinement": {
			input:       NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
			expectation: `<unknown, not null, lower bound = 1.23 (inclusive)>`,
		},
		"unknown-with-upperbound-refinement": {
			input:       NewNumberUnknown().RefineWithUpperBound(big.NewFloat(4.56), false),
			expectation: `<unknown, not null, upper bound = 4.56 (exclusive)>`,
		},
		"unknown-with-both-bound-refinements": {
			input:       NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true).RefineWithUpperBound(big.NewFloat(4.56), false),
			expectation: `<unknown, not null, lower bound = 1.23 (inclusive), upper bound = 4.56 (exclusive)>`,
		},
		"null": {
			input:       NewNumberNull(),
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

func TestNumberValueValueBigFloat(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    NumberValue
		expected *big.Float
	}{
		"known": {
			input:    NewNumberValue(big.NewFloat(2.4)),
			expected: big.NewFloat(2.4),
		},
		"known-nil": {
			input:    NewNumberValue(nil),
			expected: nil,
		},
		"null": {
			input:    NewNumberNull(),
			expected: nil,
		},
		"unknown": {
			input:    NewNumberUnknown(),
			expected: nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueBigFloat()

			if got == nil && testCase.expected != nil {
				t.Fatalf("got nil, expected: %s", testCase.expected)
			}

			if got != nil {
				if testCase.expected == nil {
					t.Fatalf("expected nil, got: %s", got)
				}

				if got.Cmp(testCase.expected) != 0 {
					t.Fatalf("expected %s, got: %s", testCase.expected, got)
				}
			}
		})
	}
}

func TestNumberValue_NotNullRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           NumberValue
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewNumberValue(big.NewFloat(4.56)).RefineAsNotNull(),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewNumberNull().RefineAsNotNull(),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewNumberUnknown(),
			expectedFound: false,
		},
		"unknown-with-notnull-refinement": {
			input:           NewNumberUnknown().RefineAsNotNull(),
			expectedRefnVal: refinement.NewNotNull(),
			expectedFound:   true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, found := test.input.NotNullRefinement()
			if found != test.expectedFound {
				t.Fatalf("Expected refinement exists to be: %t, got: %t", test.expectedFound, found)
			}

			if got == nil && test.expectedRefnVal == nil {
				// Success!
				return
			}

			if got == nil && test.expectedRefnVal != nil {
				t.Fatalf("Expected refinement data: <%+v>, got: nil", test.expectedRefnVal)
			}

			if diff := cmp.Diff(*got, test.expectedRefnVal); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNumberValue_LowerBoundRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           NumberValue
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewNumberValue(big.NewFloat(4.56)).RefineWithLowerBound(big.NewFloat(1.23), true),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewNumberNull().RefineWithLowerBound(big.NewFloat(1.23), true),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewNumberUnknown(),
			expectedFound: false,
		},
		"unknown-with-lowerbound-refinement": {
			input:           NewNumberUnknown().RefineWithLowerBound(big.NewFloat(1.23), true),
			expectedRefnVal: refinement.NewNumberLowerBound(big.NewFloat(1.23), true),
			expectedFound:   true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, found := test.input.LowerBoundRefinement()
			if found != test.expectedFound {
				t.Fatalf("Expected refinement exists to be: %t, got: %t", test.expectedFound, found)
			}

			if got == nil && test.expectedRefnVal == nil {
				// Success!
				return
			}

			if got == nil && test.expectedRefnVal != nil {
				t.Fatalf("Expected refinement data: <%+v>, got: nil", test.expectedRefnVal)
			}

			if diff := cmp.Diff(*got, test.expectedRefnVal); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNumberValue_UpperBoundRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           NumberValue
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewNumberValue(big.NewFloat(4.56)).RefineWithUpperBound(big.NewFloat(1.23), true),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewNumberNull().RefineWithUpperBound(big.NewFloat(1.23), true),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewNumberUnknown(),
			expectedFound: false,
		},
		"unknown-with-upperbound-refinement": {
			input:           NewNumberUnknown().RefineWithUpperBound(big.NewFloat(1.23), true),
			expectedRefnVal: refinement.NewNumberUpperBound(big.NewFloat(1.23), true),
			expectedFound:   true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, found := test.input.UpperBoundRefinement()
			if found != test.expectedFound {
				t.Fatalf("Expected refinement exists to be: %t, got: %t", test.expectedFound, found)
			}

			if got == nil && test.expectedRefnVal == nil {
				// Success!
				return
			}

			if got == nil && test.expectedRefnVal != nil {
				t.Fatalf("Expected refinement data: <%+v>, got: nil", test.expectedRefnVal)
			}

			if diff := cmp.Diff(*got, test.expectedRefnVal); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
