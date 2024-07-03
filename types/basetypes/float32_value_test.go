// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestFloat32ValueToTerraformValue(t *testing.T) {
	t.Parallel()

	var Float32Val float32 = 123.456

	type testCase struct {
		input       Float32Value
		expectation interface{}
	}
	tests := map[string]testCase{
		"known-int": {
			input:       NewFloat32Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"known-float": {
			input:       NewFloat32Value(Float32Val),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(float64(Float32Val))),
		},
		"unknown": {
			input:       NewFloat32Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       NewFloat32Null(),
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"deprecated-value-int": {
			input:       NewFloat32Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"deprecated-value-float": {
			input:       NewFloat32Value(Float32Val),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(float64(Float32Val))),
		},
		"deprecated-unknown": {
			input:       NewFloat32Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       NewFloat32Null(),
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

func TestFloat32ValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float32Value
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-53-precison-same": {
			input:       NewFloat32Value(123.123),
			candidate:   NewFloat32Value(123.123),
			expectation: true,
		},
		"known-known-53-precison-diff": {
			input:       NewFloat32Value(123.123),
			candidate:   NewFloat32Value(456.456),
			expectation: false,
		},
		"known-known-512-precision-same": {
			input: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			candidate: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			expectation: true,
		},
		"known-known-512-precision-diff": {
			input: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			candidate: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009"),
			},
			expectation: false,
		},
		"known-known-512-precision-mantissa-diff": {
			input: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			candidate: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.01"),
			},
			expectation: false,
		},
		"known-known-precisiondiff-mantissa-same": {
			input: NewFloat32Value(123),
			candidate: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("123"),
			},
			expectation: true,
		},
		"known-known-precisiondiff-mantissa-diff": {
			input: NewFloat32Value(0.1),
			candidate: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.1"),
			},
			expectation: false,
		},
		"known-known-precisiondiff-mantissa-same-stringdiff": {
			input: NewFloat32Value(340282346638528859811704183484516925440),
			candidate: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("340282346638528859811704183484516925440"),
			},
			expectation: false,
		},
		"known-known-zero-negativezero": {
			input:       NewFloat32Value(0),
			candidate:   NewFloat32Value(float32(math.Copysign(0, -1))),
			expectation: true,
		},
		"knownnil-known": {
			input: Float32Value{
				state: attr.ValueStateKnown,
				value: nil,
			},
			candidate:   NewFloat32Value(0.1),
			expectation: false,
		},
		"known-knownnil": {
			input: NewFloat32Value(0.1),
			candidate: Float32Value{
				state: attr.ValueStateKnown,
				value: nil,
			},
			expectation: false,
		},
		"knownnil-knownnil": {
			input: Float32Value{
				state: attr.ValueStateKnown,
				value: nil,
			},
			candidate: Float32Value{
				state: attr.ValueStateKnown,
				value: nil,
			},
			expectation: true,
		},
		"known-unknown": {
			input:       NewFloat32Value(123),
			candidate:   NewFloat32Unknown(),
			expectation: false,
		},
		"known-null": {
			input:       NewFloat32Value(123),
			candidate:   NewFloat32Null(),
			expectation: false,
		},
		"unknown-value": {
			input:       NewFloat32Unknown(),
			candidate:   NewFloat32Value(123),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NewFloat32Unknown(),
			candidate:   NewFloat32Unknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       NewFloat32Unknown(),
			candidate:   NewFloat32Null(),
			expectation: false,
		},
		"null-known": {
			input:       NewFloat32Null(),
			candidate:   NewFloat32Value(123),
			expectation: false,
		},
		"null-unknown": {
			input:       NewFloat32Null(),
			candidate:   NewFloat32Unknown(),
			expectation: false,
		},
		"null-null": {
			input:       NewFloat32Null(),
			candidate:   NewFloat32Null(),
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

func TestFloat32ValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float32Value
		expected bool
	}{
		"known": {
			input:    NewFloat32Value(2.4),
			expected: false,
		},
		"null": {
			input:    NewFloat32Null(),
			expected: true,
		},
		"unknown": {
			input:    NewFloat32Unknown(),
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

func TestFloat32ValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float32Value
		expected bool
	}{
		"known": {
			input:    NewFloat32Value(2.4),
			expected: false,
		},
		"null": {
			input:    NewFloat32Null(),
			expected: false,
		},
		"unknown": {
			input:    NewFloat32Unknown(),
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

func TestFloat32ValueString(t *testing.T) {
	t.Parallel()

	var lessThanOne float32 = 0.12340984302980000
	var moreThanOne float32 = 923879.32812
	var negativeLessThanOne float32 = -0.12340984302980000
	var negativeMoreThanOne float32 = -923879.32812
	var smallestNonZero float32 = math.SmallestNonzeroFloat32
	var largestFloat32 float32 = math.MaxFloat32

	type testCase struct {
		input       Float32Value
		expectation string
	}
	tests := map[string]testCase{
		"less-than-one": {
			input:       NewFloat32Value(lessThanOne),
			expectation: "0.123410",
		},
		"more-than-one": {
			input:       NewFloat32Value(moreThanOne),
			expectation: "923879.312500",
		},
		"negative-less-than-one": {
			input:       NewFloat32Value(negativeLessThanOne),
			expectation: "-0.123410",
		},
		"negative-more-than-one": {
			input:       NewFloat32Value(negativeMoreThanOne),
			expectation: "-923879.312500",
		},
		"min-float32": {
			input:       NewFloat32Value(smallestNonZero),
			expectation: "0.000000",
		},
		"max-float32": {
			input:       NewFloat32Value(largestFloat32),
			expectation: "340282346638528859811704183484516925440.000000",
		},
		"unknown": {
			input:       NewFloat32Unknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewFloat32Null(),
			expectation: "<null>",
		},
		"zero-value": {
			input:       Float32Value{},
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

func TestFloat32ValueValueFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float32Value
		expected float32
	}{
		"known": {
			input:    NewFloat32Value(2.4),
			expected: 2.4,
		},
		"null": {
			input:    NewFloat32Null(),
			expected: 0.0,
		},
		"unknown": {
			input:    NewFloat32Unknown(),
			expected: 0.0,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueFloat32()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat32ValueValueFloat32Pointer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float32Value
		expected *float32
	}{
		"known": {
			input:    NewFloat32Value(2.4),
			expected: pointer(float32(2.4)),
		},
		"null": {
			input:    NewFloat32Null(),
			expected: nil,
		},
		"unknown": {
			input:    NewFloat32Unknown(),
			expected: pointer(float32(0.0)),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueFloat32Pointer()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNewFloat32PointerValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value    *float32
		expected Float32Value
	}{
		"nil": {
			value:    nil,
			expected: NewFloat32Null(),
		},
		"value": {
			value:    pointer(float32(1.2)),
			expected: NewFloat32Value(1.2),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := NewFloat32PointerValue(testCase.value)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat32ValueFloat32SemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentFloat32 Float32Value
		givenFloat32   Float32Value
		expectedMatch  bool
		expectedDiags  diag.Diagnostics
	}{
		"not equal - whole number": {
			currentFloat32: NewFloat32Value(1),
			givenFloat32:   NewFloat32Value(2),
			expectedMatch:  false,
		},
		"not equal - float": {
			currentFloat32: NewFloat32Value(1.1),
			givenFloat32:   NewFloat32Value(1.2),
			expectedMatch:  false,
		},
		"not equal - float differing precision": {
			currentFloat32: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.01"),
			},
			givenFloat32:  NewFloat32Value(0.02),
			expectedMatch: false,
		},
		"semantically equal - whole number": {
			currentFloat32: NewFloat32Value(1),
			givenFloat32:   NewFloat32Value(1),
			expectedMatch:  true,
		},
		"semantically equal - float": {
			currentFloat32: NewFloat32Value(1.1),
			givenFloat32:   NewFloat32Value(1.1),
			expectedMatch:  true,
		},
		"semantically equal - float differing precision": {
			currentFloat32: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.01"),
			},
			givenFloat32:  NewFloat32Value(0.01),
			expectedMatch: true,
		},
		// Only 53 bits of precision are compared, Go built-in float32
		"semantically equal - float 512 precision, different value not significant": {
			currentFloat32: Float32Value{
				state: attr.ValueStateKnown,
				value: testMustParseFloat("0.010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"),
			},
			givenFloat32: Float32Value{
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

			match, diags := testCase.currentFloat32.Float32SemanticEquals(context.Background(), testCase.givenFloat32)

			if testCase.expectedMatch != match {
				t.Errorf("Expected Float32SemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
