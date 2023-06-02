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
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
		"known-known-same": {
			input:       NewFloat64Value(123),
			candidate:   NewFloat64Value(123),
			expectation: true,
		},
		"known-known-diff": {
			input:       NewFloat64Value(123),
			candidate:   NewFloat64Value(456),
			expectation: false,
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
