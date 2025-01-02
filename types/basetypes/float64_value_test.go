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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/refinement"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

// testMustParseFloat parses a string into a *big.Float similar to cty and
// tftypes logic or panics on any error.
//
// Reference: https://github.com/hashicorp/go-cty/blob/85980079f637862fa8e43ddc82dd74315e2f4c85/cty/value_init.go#L49
// Reference: https://github.com/hashicorp/terraform-plugin-go/blob/c593d2e0da8d2258b2a22af867c39842a0cb89f7/tftypes/value_msgpack.go#L108
func testMustParseFloat(s string) *big.Float {
	f, _, err := big.ParseFloat(s, 10, 512, big.ToNearestEven)

	if err != nil {
		panic(err)
	}

	return f
}

func TestFloat64ValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64Value
		expectation interface{}
	}
	tests := map[string]testCase{
		"known-int": {
			input:       NewFloat64Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"known-float": {
			input:       NewFloat64Value(123.456),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.456)),
		},
		"unknown": {
			input:       NewFloat64Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"unknown-with-notnull-refinement": {
			input: NewFloat64Unknown().RefineAsNotNull(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness: tfrefinement.NewNullness(false),
			}),
		},
		"unknown-with-lower-bound-refinement": {
			input: NewFloat64Unknown().RefineWithLowerBound(1.23, true),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(1.23), true),
			}),
		},
		"unknown-with-upper-bound-refinement": {
			input: NewFloat64Unknown().RefineWithUpperBound(4.56, false),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(4.56), false),
			}),
		},
		"unknown-with-both-bound-refinements": {
			input: NewFloat64Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.56, false),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:         tfrefinement.NewNullness(false),
				tfrefinement.KeyNumberLowerBound: tfrefinement.NewNumberLowerBound(big.NewFloat(1.23), true),
				tfrefinement.KeyNumberUpperBound: tfrefinement.NewNumberUpperBound(big.NewFloat(4.56), false),
			}),
		},
		"null": {
			input:       NewFloat64Null(),
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"deprecated-value-int": {
			input:       NewFloat64Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"deprecated-value-float": {
			input:       NewFloat64Value(123.456),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.456)),
		},
		"deprecated-unknown": {
			input:       NewFloat64Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       NewFloat64Null(),
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

func TestFloat64ValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64Value
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-53-precison-same": {
			input:       NewFloat64Value(123.123),
			candidate:   NewFloat64Value(123.123),
			expectation: true,
		},
		"known-known-53-precison-diff": {
			input:       NewFloat64Value(123.123),
			candidate:   NewFloat64Value(456.456),
			expectation: false,
		},
		"known-known-512-precision-same": {
			input: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			candidate: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			expectation: true,
		},
		"known-known-512-precision-diff": {
			input: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			candidate: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009"),
			},
			expectation: false,
		},
		"known-known-512-precision-mantissa-diff": {
			input: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			candidate: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.01"),
			},
			expectation: false,
		},
		"known-known-precisiondiff-mantissa-same": {
			input: NewFloat64Value(123),
			candidate: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("123"),
			},
			expectation: true,
		},
		"known-known-precisiondiff-mantissa-diff": {
			input: NewFloat64Value(0.1),
			candidate: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.1"),
			},
			expectation: false,
		},
		"knownnil-known": {
			input: Float64Value{
				state: attr.ValueStateKnown,
				value: nil,
			},
			candidate:   NewFloat64Value(0.1),
			expectation: false,
		},
		"known-knownnil": {
			input: NewFloat64Value(0.1),
			candidate: Float64Value{
				state: attr.ValueStateKnown,
				value: nil,
			},
			expectation: false,
		},
		"knownnil-knownnil": {
			input: Float64Value{
				state: attr.ValueStateKnown,
				value: nil,
			},
			candidate: Float64Value{
				state: attr.ValueStateKnown,
				value: nil,
			},
			expectation: true,
		},
		"known-unknown": {
			input:       NewFloat64Value(123),
			candidate:   NewFloat64Unknown(),
			expectation: false,
		},
		"known-null": {
			input:       NewFloat64Value(123),
			candidate:   NewFloat64Null(),
			expectation: false,
		},
		"unknown-value": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Value(123),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Unknown(),
			expectation: true,
		},
		"unknown-unknown-with-notnull-refinement": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Unknown().RefineAsNotNull(),
			expectation: false,
		},
		"unknown-unknown-with-lowerbound-refinement": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Unknown().RefineWithLowerBound(1.23, true),
			expectation: false,
		},
		"unknown-unknown-with-upperbound-refinement": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Unknown().RefineWithUpperBound(4.56, false),
			expectation: false,
		},
		"unknowns-with-matching-notnull-refinements": {
			input:       NewFloat64Unknown().RefineAsNotNull(),
			candidate:   NewFloat64Unknown().RefineAsNotNull(),
			expectation: true,
		},
		"unknowns-with-matching-lowerbound-refinements": {
			input:       NewFloat64Unknown().RefineWithLowerBound(1.23, true),
			candidate:   NewFloat64Unknown().RefineWithLowerBound(1.23, true),
			expectation: true,
		},
		"unknowns-with-different-lowerbound-refinements": {
			input:       NewFloat64Unknown().RefineWithLowerBound(1.23, true),
			candidate:   NewFloat64Unknown().RefineWithLowerBound(1.24, true),
			expectation: false,
		},
		"unknowns-with-different-lowerbound-refinements-inclusive": {
			input:       NewFloat64Unknown().RefineWithLowerBound(1.23, true),
			candidate:   NewFloat64Unknown().RefineWithLowerBound(1.23, false),
			expectation: false,
		},
		"unknowns-with-matching-upperbound-refinements": {
			input:       NewFloat64Unknown().RefineWithUpperBound(4.56, true),
			candidate:   NewFloat64Unknown().RefineWithUpperBound(4.56, true),
			expectation: true,
		},
		"unknowns-with-different-upperbound-refinements": {
			input:       NewFloat64Unknown().RefineWithUpperBound(4.56, true),
			candidate:   NewFloat64Unknown().RefineWithUpperBound(4.57, true),
			expectation: false,
		},
		"unknowns-with-different-upperbound-refinements-inclusive": {
			input:       NewFloat64Unknown().RefineWithUpperBound(4.56, true),
			candidate:   NewFloat64Unknown().RefineWithUpperBound(4.56, false),
			expectation: false,
		},
		"unknowns-with-matching-both-bound-refinements": {
			input:       NewFloat64Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.56, true),
			candidate:   NewFloat64Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.56, true),
			expectation: true,
		},
		"unknowns-with-different-both-bound-refinements": {
			input:       NewFloat64Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.56, true),
			candidate:   NewFloat64Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.57, true),
			expectation: false,
		},
		"unknowns-with-different-both-bound-refinements-inclusive": {
			input:       NewFloat64Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.56, true),
			candidate:   NewFloat64Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.56, false),
			expectation: false,
		},
		"unknown-null": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Null(),
			expectation: false,
		},
		"null-known": {
			input:       NewFloat64Null(),
			candidate:   NewFloat64Value(123),
			expectation: false,
		},
		"null-unknown": {
			input:       NewFloat64Null(),
			candidate:   NewFloat64Unknown(),
			expectation: false,
		},
		"null-null": {
			input:       NewFloat64Null(),
			candidate:   NewFloat64Null(),
			expectation: true,
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

func TestFloat64ValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64Value
		expected bool
	}{
		"known": {
			input:    NewFloat64Value(2.4),
			expected: false,
		},
		"null": {
			input:    NewFloat64Null(),
			expected: true,
		},
		"unknown": {
			input:    NewFloat64Unknown(),
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

func TestFloat64ValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64Value
		expected bool
	}{
		"known": {
			input:    NewFloat64Value(2.4),
			expected: false,
		},
		"null": {
			input:    NewFloat64Null(),
			expected: false,
		},
		"unknown": {
			input:    NewFloat64Unknown(),
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

func TestFloat64ValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64Value
		expectation string
	}
	tests := map[string]testCase{
		"less-than-one": {
			input:       NewFloat64Value(0.12340984302980000),
			expectation: "0.123410",
		},
		"more-than-one": {
			input:       NewFloat64Value(92387938173219.327663),
			expectation: "92387938173219.328125",
		},
		"negative-more-than-one": {
			input:       NewFloat64Value(-0.12340984302980000),
			expectation: "-0.123410",
		},
		"negative-less-than-one": {
			input:       NewFloat64Value(-92387938173219.327663),
			expectation: "-92387938173219.328125",
		},
		"min-float64": {
			input:       NewFloat64Value(math.SmallestNonzeroFloat64),
			expectation: "0.000000",
		},
		"max-float64": {
			input:       NewFloat64Value(math.MaxFloat64),
			expectation: "179769313486231570814527423731704356798070567525844996598917476803157260780028538760589558632766878171540458953514382464234321326889464182768467546703537516986049910576551282076245490090389328944075868508455133942304583236903222948165808559332123348274797826204144723168738177180919299881250404026184124858368.000000",
		},
		"unknown": {
			input:       NewFloat64Unknown(),
			expectation: "<unknown>",
		},
		"unknown-with-notnull-refinement": {
			input:       NewFloat64Unknown().RefineAsNotNull(),
			expectation: "<unknown, not null>",
		},
		"unknown-with-lowerbound-refinement": {
			input:       NewFloat64Unknown().RefineWithLowerBound(1.23, true),
			expectation: `<unknown, not null, lower bound = 1.230000 (inclusive)>`,
		},
		"unknown-with-upperbound-refinement": {
			input:       NewFloat64Unknown().RefineWithUpperBound(4.56, false),
			expectation: `<unknown, not null, upper bound = 4.560000 (exclusive)>`,
		},
		"unknown-with-both-bound-refinements": {
			input:       NewFloat64Unknown().RefineWithLowerBound(1.23, true).RefineWithUpperBound(4.56, false),
			expectation: `<unknown, not null, lower bound = 1.230000 (inclusive), upper bound = 4.560000 (exclusive)>`,
		},
		"null": {
			input:       NewFloat64Null(),
			expectation: "<null>",
		},
		"zero-value": {
			input:       Float64Value{},
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

func TestFloat64ValueValueFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64Value
		expected float64
	}{
		"known": {
			input:    NewFloat64Value(2.4),
			expected: 2.4,
		},
		"null": {
			input:    NewFloat64Null(),
			expected: 0.0,
		},
		"unknown": {
			input:    NewFloat64Unknown(),
			expected: 0.0,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueFloat64()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64ValueValueFloat64Pointer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64Value
		expected *float64
	}{
		"known": {
			input:    NewFloat64Value(2.4),
			expected: pointer(2.4),
		},
		"null": {
			input:    NewFloat64Null(),
			expected: nil,
		},
		"unknown": {
			input:    NewFloat64Unknown(),
			expected: pointer(0.0),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueFloat64Pointer()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNewFloat64PointerValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value    *float64
		expected Float64Value
	}{
		"nil": {
			value:    nil,
			expected: NewFloat64Null(),
		},
		"value": {
			value:    pointer(1.2),
			expected: NewFloat64Value(1.2),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := NewFloat64PointerValue(testCase.value)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64ValueFloat64SemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentFloat64 Float64Value
		givenFloat64   Float64Value
		expectedMatch  bool
		expectedDiags  diag.Diagnostics
	}{
		"not equal - whole number": {
			currentFloat64: NewFloat64Value(1),
			givenFloat64:   NewFloat64Value(2),
			expectedMatch:  false,
		},
		"not equal - float": {
			currentFloat64: NewFloat64Value(1.1),
			givenFloat64:   NewFloat64Value(1.2),
			expectedMatch:  false,
		},
		"not equal - float differing precision": {
			currentFloat64: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.01"),
			},
			givenFloat64:  NewFloat64Value(0.02),
			expectedMatch: false,
		},
		"semantically equal - whole number": {
			currentFloat64: NewFloat64Value(1),
			givenFloat64:   NewFloat64Value(1),
			expectedMatch:  true,
		},
		"semantically equal - float": {
			currentFloat64: NewFloat64Value(1.1),
			givenFloat64:   NewFloat64Value(1.1),
			expectedMatch:  true,
		},
		"semantically equal - float differing precision": {
			currentFloat64: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.01"),
			},
			givenFloat64:  NewFloat64Value(0.01),
			expectedMatch: true,
		},
		// Only 53 bits of precision are compared, Go built-in float64
		"semantically equal - float 512 precision, different value not significant": {
			currentFloat64: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			givenFloat64: Float64Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009"),
			},
			expectedMatch: true,
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			match, diags := testCase.currentFloat64.Float64SemanticEquals(context.Background(), testCase.givenFloat64)

			if testCase.expectedMatch != match {
				t.Errorf("Expected Float64SemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestFloat64Value_NotNullRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           Float64Value
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewFloat64Value(4.56).RefineAsNotNull(),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewFloat64Null().RefineAsNotNull(),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewFloat64Unknown(),
			expectedFound: false,
		},
		"unknown-with-notnull-refinement": {
			input:           NewFloat64Unknown().RefineAsNotNull(),
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

func TestFloat64Value_LowerBoundRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           Float64Value
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewFloat64Value(4.56).RefineWithLowerBound(1.23, true),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewFloat64Null().RefineWithLowerBound(1.23, true),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewFloat64Unknown(),
			expectedFound: false,
		},
		"unknown-with-lowerbound-refinement": {
			input:           NewFloat64Unknown().RefineWithLowerBound(1.23, true),
			expectedRefnVal: refinement.NewFloat64LowerBound(1.23, true),
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

func TestFloat64Value_UpperBoundRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           Float64Value
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewFloat64Value(4.56).RefineWithUpperBound(1.23, true),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewFloat64Null().RefineWithUpperBound(1.23, true),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewFloat64Unknown(),
			expectedFound: false,
		},
		"unknown-with-upperbound-refinement": {
			input:           NewFloat64Unknown().RefineWithUpperBound(1.23, true),
			expectedRefnVal: refinement.NewFloat64UpperBound(1.23, true),
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
