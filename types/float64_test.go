package types

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestFloat64ValueDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	knownFloat64 := Float64Value(2.4)

	knownFloat64.Null = true

	if knownFloat64.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	knownFloat64.Unknown = true

	if knownFloat64.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	knownFloat64.Value = 4.8

	if knownFloat64.ValueFloat64() == 4.8 {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestFloat64NullDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	nullFloat64 := Float64Null()

	nullFloat64.Null = false

	if !nullFloat64.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	nullFloat64.Unknown = true

	if nullFloat64.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	nullFloat64.Value = 4.8

	if nullFloat64.ValueFloat64() == 4.8 {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestFloat64UnknownDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	unknownFloat64 := Float64Unknown()

	unknownFloat64.Null = true

	if unknownFloat64.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	unknownFloat64.Unknown = false

	if !unknownFloat64.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	unknownFloat64.Value = 4.8

	if unknownFloat64.ValueFloat64() == 4.8 {
		t.Error("unexpected value update after Value field setting")
	}
}

func TestFloat64ValueFromTerraform(t *testing.T) {
	t.Parallel()

	testFloat64ValueFromTerraform(t, true)
}

func testFloat64ValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value-int": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: Float64{Value: 123.0},
		},
		"value-float": {
			input:       tftypes.NewValue(tftypes.Number, 123.456),
			expectation: Float64{Value: 123.456},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: Float64{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: Float64{Null: true},
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

			f := Float64Type.ValueFromTerraform
			if direct {
				f = float64ValueFromTerraform
			}
			got, err := f(ctx, test.input)
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
			if err == nil && test.expectedErr != "" {
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

func TestFloat64ToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64
		expectation interface{}
	}
	tests := map[string]testCase{
		"known-int": {
			input:       Float64Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"known-float": {
			input:       Float64Value(123.456),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.456)),
		},
		"unknown": {
			input:       Float64Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       Float64Null(),
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"deprecated-value-int": {
			input:       Float64{Value: 123},
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"deprecated-value-float": {
			input:       Float64{Value: 123.456},
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.456)),
		},
		"deprecated-unknown": {
			input:       Float64{Unknown: true},
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       Float64{Null: true},
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

func TestFloat64Equal(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       Float64Value(123),
			candidate:   Float64Value(123),
			expectation: true,
		},
		"known-known-diff": {
			input:       Float64Value(123),
			candidate:   Float64Value(456),
			expectation: false,
		},
		"known-unknown": {
			input:       Float64Value(123),
			candidate:   Float64Unknown(),
			expectation: false,
		},
		"known-null": {
			input:       Float64Value(123),
			candidate:   Float64Null(),
			expectation: false,
		},
		"unknown-value": {
			input:       Float64Unknown(),
			candidate:   Float64Value(123),
			expectation: false,
		},
		"unknown-unknown": {
			input:       Float64Unknown(),
			candidate:   Float64Unknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       Float64Unknown(),
			candidate:   Float64Null(),
			expectation: false,
		},
		"null-known": {
			input:       Float64Null(),
			candidate:   Float64Value(123),
			expectation: false,
		},
		"null-unknown": {
			input:       Float64Null(),
			candidate:   Float64Unknown(),
			expectation: false,
		},
		"null-null": {
			input:       Float64Null(),
			candidate:   Float64Null(),
			expectation: true,
		},
		"deprecated-known-known-same": {
			input:       Float64{Value: 123},
			candidate:   Float64Value(123),
			expectation: false, // intentional
		},
		"deprecated-known-known-diff": {
			input:       Float64{Value: 123},
			candidate:   Float64Value(456),
			expectation: false,
		},
		"deprecated-known-unknown": {
			input:       Float64{Value: 123},
			candidate:   Float64Unknown(),
			expectation: false,
		},
		"deprecated-known-null": {
			input:       Float64{Value: 123},
			candidate:   Float64Null(),
			expectation: false,
		},
		"deprecated-known-deprecated-known-same": {
			input:       Float64{Value: 123},
			candidate:   Float64{Value: 123},
			expectation: true,
		},
		"deprecated-known-deprecated-known-diff": {
			input:       Float64{Value: 123},
			candidate:   Float64{Value: 456},
			expectation: false,
		},
		"deprecated-known-deprecated-unknown": {
			input:       Float64{Value: 123},
			candidate:   Float64{Unknown: true},
			expectation: false,
		},
		"deprecated-known-deprecated-null": {
			input:       Float64{Value: 123},
			candidate:   Float64{Null: true},
			expectation: false,
		},
		"deprecated-known-wrongType": {
			input:       Float64{Value: 123},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-known-nil": {
			input:       Float64{Value: 123},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-unknown-value": {
			input:       Float64{Unknown: true},
			candidate:   Float64Value(123),
			expectation: false,
		},
		"deprecated-unknown-unknown": {
			input:       Float64{Unknown: true},
			candidate:   Float64Unknown(),
			expectation: false, // intentional
		},
		"deprecated-unknown-null": {
			input:       Float64{Unknown: true},
			candidate:   Float64Null(),
			expectation: false,
		},
		"deprecated-unknown-deprecated-value": {
			input:       Float64{Unknown: true},
			candidate:   Float64{Value: 123},
			expectation: false,
		},
		"deprecated-unknown-deprecated-unknown": {
			input:       Float64{Unknown: true},
			candidate:   Float64{Unknown: true},
			expectation: true,
		},
		"deprecated-unknown-deprecated-null": {
			input:       Float64{Unknown: true},
			candidate:   Float64{Null: true},
			expectation: false,
		},
		"deprecated-unknown-wrongType": {
			input:       Float64{Unknown: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-unknown-nil": {
			input:       Float64{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-null-known": {
			input:       Float64{Null: true},
			candidate:   Float64Value(123),
			expectation: false,
		},
		"deprecated-null-unknown": {
			input:       Float64{Null: true},
			candidate:   Float64Unknown(),
			expectation: false,
		},
		"deprecated-null-null": {
			input:       Float64{Null: true},
			candidate:   Float64Null(),
			expectation: false, // intentional
		},
		"deprecated-null-deprecated-known": {
			input:       Float64{Null: true},
			candidate:   Float64{Value: 123},
			expectation: false,
		},
		"deprecated-null-deprecated-unknown": {
			input:       Float64{Null: true},
			candidate:   Float64{Unknown: true},
			expectation: false,
		},
		"deprecated-null-deprecated-null": {
			input:       Float64{Null: true},
			candidate:   Float64{Null: true},
			expectation: true,
		},
		"deprecated-null-wrongType": {
			input:       Float64{Null: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-null-nil": {
			input:       Float64{Null: true},
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

func TestFloat64IsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64
		expected bool
	}{
		"known": {
			input:    Float64Value(2.4),
			expected: false,
		},
		"deprecated-known": {
			input:    Float64{Value: 2.4},
			expected: false,
		},
		"null": {
			input:    Float64Null(),
			expected: true,
		},
		"deprecated-null": {
			input:    Float64{Null: true},
			expected: true,
		},
		"unknown": {
			input:    Float64Unknown(),
			expected: false,
		},
		"deprecated-unknown": {
			input:    Float64{Unknown: true},
			expected: false,
		},
		"deprecated-invalid": {
			input:    Float64{Null: true, Unknown: true},
			expected: true,
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

func TestFloat64IsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64
		expected bool
	}{
		"known": {
			input:    Float64Value(2.4),
			expected: false,
		},
		"deprecated-known": {
			input:    Float64{Value: 2.4},
			expected: false,
		},
		"null": {
			input:    Float64Null(),
			expected: false,
		},
		"deprecated-null": {
			input:    Float64{Null: true},
			expected: false,
		},
		"unknown": {
			input:    Float64Unknown(),
			expected: true,
		},
		"deprecated-unknown": {
			input:    Float64{Unknown: true},
			expected: true,
		},
		"deprecated-invalid": {
			input:    Float64{Null: true, Unknown: true},
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

func TestFloat64String(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64
		expectation string
	}
	tests := map[string]testCase{
		"less-than-one": {
			input:       Float64Value(0.12340984302980000),
			expectation: "0.123410",
		},
		"more-than-one": {
			input:       Float64Value(92387938173219.327663),
			expectation: "92387938173219.328125",
		},
		"negative-more-than-one": {
			input:       Float64Value(-0.12340984302980000),
			expectation: "-0.123410",
		},
		"negative-less-than-one": {
			input:       Float64Value(-92387938173219.327663),
			expectation: "-92387938173219.328125",
		},
		"min-float64": {
			input:       Float64Value(math.SmallestNonzeroFloat64),
			expectation: "0.000000",
		},
		"max-float64": {
			input:       Float64Value(math.MaxFloat64),
			expectation: "179769313486231570814527423731704356798070567525844996598917476803157260780028538760589558632766878171540458953514382464234321326889464182768467546703537516986049910576551282076245490090389328944075868508455133942304583236903222948165808559332123348274797826204144723168738177180919299881250404026184124858368.000000",
		},
		"unknown": {
			input:       Float64Unknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       Float64Null(),
			expectation: "<null>",
		},
		"deprecated-known-less-than-one": {
			input:       Float64{Value: 0.12340984302980000},
			expectation: "0.123410",
		},
		"deprecated-known-more-than-one": {
			input:       Float64{Value: 92387938173219.327663},
			expectation: "92387938173219.328125",
		},
		"deprecated-known-negative-more-than-one": {
			input:       Float64{Value: -0.12340984302980000},
			expectation: "-0.123410",
		},
		"deprecated-known-negative-less-than-one": {
			input:       Float64{Value: -92387938173219.327663},
			expectation: "-92387938173219.328125",
		},
		"deprecated-known-min-float64": {
			input:       Float64{Value: math.SmallestNonzeroFloat64},
			expectation: "0.000000",
		},
		"deprecated-known-max-float64": {
			input:       Float64{Value: math.MaxFloat64},
			expectation: "179769313486231570814527423731704356798070567525844996598917476803157260780028538760589558632766878171540458953514382464234321326889464182768467546703537516986049910576551282076245490090389328944075868508455133942304583236903222948165808559332123348274797826204144723168738177180919299881250404026184124858368.000000",
		},
		"deprecated-known-unknown": {
			input:       Float64{Unknown: true},
			expectation: "<unknown>",
		},
		"deprecated-known-null": {
			input:       Float64{Null: true},
			expectation: "<null>",
		},
		"default-0": {
			input:       Float64{},
			expectation: "0.000000",
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

func TestFloat64ValueFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64
		expected float64
	}{
		"known": {
			input:    Float64Value(2.4),
			expected: 2.4,
		},
		"deprecated-known": {
			input:    Float64{Value: 2.4},
			expected: 2.4,
		},
		"null": {
			input:    Float64Null(),
			expected: 0.0,
		},
		"deprecated-null": {
			input:    Float64{Null: true},
			expected: 0.0,
		},
		"unknown": {
			input:    Float64Unknown(),
			expected: 0.0,
		},
		"deprecated-unknown": {
			input:    Float64{Unknown: true},
			expected: 0.0,
		},
		"deprecated-invalid": {
			input:    Float64{Null: true, Unknown: true},
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
