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

func numberComparer(i, j *big.Float) bool {
	return (i == nil && j == nil) || (i != nil && j != nil && i.Cmp(j) == 0)
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestNumberValueDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	knownNumber := NumberValue(big.NewFloat(2.4))

	knownNumber.Null = true

	if knownNumber.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	knownNumber.Unknown = true

	if knownNumber.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	knownNumber.Value = big.NewFloat(4.8)

	if knownNumber.ValueBigFloat().Cmp(big.NewFloat(4.8)) == 0 {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestNumberNullDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	nullNumber := NumberNull()

	nullNumber.Null = false

	if !nullNumber.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	nullNumber.Unknown = true

	if nullNumber.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	nullNumber.Value = big.NewFloat(4.8)

	if nullNumber.ValueBigFloat() != nil {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestNumberUnknownDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	unknownNumber := NumberUnknown()

	unknownNumber.Null = true

	if unknownNumber.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	unknownNumber.Unknown = false

	if !unknownNumber.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	unknownNumber.Value = big.NewFloat(4.8)

	if unknownNumber.ValueBigFloat() != nil {
		t.Error("unexpected value update after Value field setting")
	}
}

func TestNumberValueFromTerraform(t *testing.T) {
	t.Parallel()

	testNumberValueFromTerraform(t, true)
}

func testNumberValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: Number{Value: big.NewFloat(123)},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: Number{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: Number{Null: true},
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

			f := NumberType.ValueFromTerraform
			if direct {
				f = numberValueFromTerraform
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

func TestNumberToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Number
		expectation tftypes.Value
	}
	tests := map[string]testCase{
		"value": {
			input:       NumberValue(big.NewFloat(123)),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
		},
		"known-nil": {
			input:       NumberValue(nil),
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"unknown": {
			input:       NumberUnknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       NumberNull(),
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"deprecated-value": {
			input:       Number{Value: big.NewFloat(123)},
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
		},
		"deprecated-known-nil": {
			input:       Number{Value: nil},
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"deprecated-unknown": {
			input:       Number{Unknown: true},
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       Number{Null: true},
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

func TestNumberEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Number
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       NumberValue(big.NewFloat(123)),
			candidate:   NumberValue(big.NewFloat(123)),
			expectation: true,
		},
		"known-known-diff": {
			input:       NumberValue(big.NewFloat(123)),
			candidate:   NumberValue(big.NewFloat(456)),
			expectation: false,
		},
		"known-nil-known": {
			input:       NumberValue(nil),
			candidate:   NumberValue(big.NewFloat(456)),
			expectation: false,
		},
		"known-nil-null": {
			input:       NumberValue(nil),
			candidate:   NumberNull(),
			expectation: true,
		},
		"known-unknown": {
			input:       NumberValue(big.NewFloat(123)),
			candidate:   NumberUnknown(),
			expectation: false,
		},
		"known-null": {
			input:       NumberValue(big.NewFloat(123)),
			candidate:   NumberNull(),
			expectation: false,
		},
		"known-wrong-type": {
			input:       NumberValue(big.NewFloat(123)),
			candidate:   Float64Value(123),
			expectation: false,
		},
		"known-nil": {
			input:       NumberValue(big.NewFloat(123)),
			candidate:   nil,
			expectation: false,
		},
		"unknown-known": {
			input:       NumberUnknown(),
			candidate:   NumberValue(big.NewFloat(123)),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NumberUnknown(),
			candidate:   NumberUnknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       NumberUnknown(),
			candidate:   NumberNull(),
			expectation: false,
		},
		"unknown-wrong-type": {
			input:       NumberUnknown(),
			candidate:   Float64Unknown(),
			expectation: false,
		},
		"unknown-nil": {
			input:       NumberUnknown(),
			candidate:   nil,
			expectation: false,
		},
		"null-known": {
			input:       NumberNull(),
			candidate:   NumberValue(big.NewFloat(123)),
			expectation: false,
		},
		"null-known-nil": {
			input:       NumberNull(),
			candidate:   NumberValue(nil),
			expectation: true,
		},
		"null-unknown": {
			input:       NumberNull(),
			candidate:   NumberUnknown(),
			expectation: false,
		},
		"null-null": {
			input:       NumberNull(),
			candidate:   NumberNull(),
			expectation: true,
		},
		"null-wrong-type": {
			input:       NumberNull(),
			candidate:   Float64Null(),
			expectation: false,
		},
		"null-nil": {
			input:       NumberNull(),
			candidate:   nil,
			expectation: false,
		},
		"deprecated-known-known-same": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   NumberValue(big.NewFloat(123)),
			expectation: false, // intentional
		},
		"deprecated-known-known-diff": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   NumberValue(big.NewFloat(456)),
			expectation: false,
		},
		"deprecated-known-unknown": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   NumberNull(),
			expectation: false,
		},
		"deprecated-known-null": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   NumberNull(),
			expectation: false,
		},
		"deprecated-known-deprecated-known-same": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: true,
		},
		"deprecated-known-deprecated-known-diff": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Value: big.NewFloat(456)},
			expectation: false,
		},
		"deprecated-known-deprecated-unknown": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Unknown: true},
			expectation: false,
		},
		"deprecated-known-deprecated-null": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Null: true},
			expectation: false,
		},
		"deprecated-known-wrongType": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-known-nil": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-known-nilValue": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Value: nil},
			expectation: false,
		},
		"deprecated-unknown-known": {
			input:       Number{Unknown: true},
			candidate:   NumberValue(big.NewFloat(123)),
			expectation: false,
		},
		"deprecated-unknown-unknown": {
			input:       Number{Unknown: true},
			candidate:   NumberUnknown(),
			expectation: false, // intentional
		},
		"deprecated-unknown-null": {
			input:       Number{Unknown: true},
			candidate:   NumberNull(),
			expectation: false,
		},
		"deprecated-unknown-deprecated-known": {
			input:       Number{Unknown: true},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"deprecated-unknown-deprecated-unknown": {
			input:       Number{Unknown: true},
			candidate:   Number{Unknown: true},
			expectation: true,
		},
		"deprecated-unknown-deprecated-null": {
			input:       Number{Unknown: true},
			candidate:   Number{Null: true},
			expectation: false,
		},
		"deprecated-unknown-wrongType": {
			input:       Number{Unknown: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-unknown-nil": {
			input:       Number{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-unknown-nilValue": {
			input:       Number{Unknown: true},
			candidate:   Number{Value: nil},
			expectation: false,
		},
		"deprecated-null-value": {
			input:       Number{Null: true},
			candidate:   NumberValue(big.NewFloat(123)),
			expectation: false,
		},
		"deprecated-null-unknown": {
			input:       Number{Null: true},
			candidate:   NumberUnknown(),
			expectation: false,
		},
		"deprecated-null-null": {
			input:       Number{Null: true},
			candidate:   NumberNull(),
			expectation: false, // intentional
		},
		"deprecated-null-deprecated-value": {
			input:       Number{Null: true},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"deprecated-null-deprecated-unknown": {
			input:       Number{Null: true},
			candidate:   Number{Unknown: true},
			expectation: false,
		},
		"deprecated-null-deprecated-null": {
			input:       Number{Null: true},
			candidate:   Number{Null: true},
			expectation: true,
		},
		"deprecated-null-wrongType": {
			input:       Number{Null: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-null-nil": {
			input:       Number{Null: true},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-null-nilValue": {
			input:       Number{Null: true},
			candidate:   Number{Value: nil},
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

func TestNumberIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Number
		expected bool
	}{
		"known": {
			input:    NumberValue(big.NewFloat(2.4)),
			expected: false,
		},
		"deprecated-known": {
			input:    Number{Value: big.NewFloat(2.4)},
			expected: false,
		},
		"null": {
			input:    NumberNull(),
			expected: true,
		},
		"deprecated-null": {
			input:    Number{Null: true},
			expected: true,
		},
		"unknown": {
			input:    NumberUnknown(),
			expected: false,
		},
		"deprecated-unknown": {
			input:    Number{Unknown: true},
			expected: false,
		},
		"deprecated-invalid": {
			input:    Number{Null: true, Unknown: true},
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

func TestNumberIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Number
		expected bool
	}{
		"known": {
			input:    NumberValue(big.NewFloat(2.4)),
			expected: false,
		},
		"deprecated-known": {
			input:    Number{Value: big.NewFloat(2.4)},
			expected: false,
		},
		"null": {
			input:    NumberNull(),
			expected: false,
		},
		"deprecated-null": {
			input:    Number{Null: true},
			expected: false,
		},
		"unknown": {
			input:    NumberUnknown(),
			expected: true,
		},
		"deprecated-unknown": {
			input:    Number{Unknown: true},
			expected: true,
		},
		"deprecated-invalid": {
			input:    Number{Null: true, Unknown: true},
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

func TestNumberString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Number
		expectation string
	}
	tests := map[string]testCase{
		"known-less-than-one": {
			input:       NumberValue(big.NewFloat(0.12340984302980000)),
			expectation: "0.123409843",
		},
		"known-more-than-one": {
			input:       NumberValue(big.NewFloat(92387938173219.327663)),
			expectation: "9.238793817e+13",
		},
		"known-negative-more-than-one": {
			input:       NumberValue(big.NewFloat(-0.12340984302980000)),
			expectation: "-0.123409843",
		},
		"known-negative-less-than-one": {
			input:       NumberValue(big.NewFloat(-92387938173219.327663)),
			expectation: "-9.238793817e+13",
		},
		"known-min-float64": {
			input:       NumberValue(big.NewFloat(math.SmallestNonzeroFloat64)),
			expectation: "4.940656458e-324",
		},
		"known-max-float64": {
			input:       NumberValue(big.NewFloat(math.MaxFloat64)),
			expectation: "1.797693135e+308",
		},
		"unknown": {
			input:       NumberUnknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       NumberNull(),
			expectation: "<null>",
		},
		"deprecated-known-less-than-one": {
			input:       Number{Value: big.NewFloat(0.12340984302980000)},
			expectation: "0.123409843",
		},
		"deprecated-known-more-than-one": {
			input:       Number{Value: big.NewFloat(92387938173219.327663)},
			expectation: "9.238793817e+13",
		},
		"deprecated-known-negative-more-than-one": {
			input:       Number{Value: big.NewFloat(-0.12340984302980000)},
			expectation: "-0.123409843",
		},
		"deprecated-known-negative-less-than-one": {
			input:       Number{Value: big.NewFloat(-92387938173219.327663)},
			expectation: "-9.238793817e+13",
		},
		"deprecated-known-min-float64": {
			input:       Number{Value: big.NewFloat(math.SmallestNonzeroFloat64)},
			expectation: "4.940656458e-324",
		},
		"deprecated-known-max-float64": {
			input:       Number{Value: big.NewFloat(math.MaxFloat64)},
			expectation: "1.797693135e+308",
		},
		"deprecated-unknown": {
			input:       Number{Unknown: true},
			expectation: "<unknown>",
		},
		"deprecated-null": {
			input:       Number{Null: true},
			expectation: "<null>",
		},
		"default-null": {
			input:       Number{},
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

func TestNumberValueBigFloat(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Number
		expected *big.Float
	}{
		"known": {
			input:    NumberValue(big.NewFloat(2.4)),
			expected: big.NewFloat(2.4),
		},
		"known-nil": {
			input:    NumberValue(nil),
			expected: nil,
		},
		"deprecated-known": {
			input:    Number{Value: big.NewFloat(2.4)},
			expected: big.NewFloat(2.4),
		},
		"null": {
			input:    NumberNull(),
			expected: nil,
		},
		"deprecated-null": {
			input:    Number{Null: true},
			expected: nil,
		},
		"unknown": {
			input:    NumberUnknown(),
			expected: nil,
		},
		"deprecated-unknown": {
			input:    Number{Unknown: true},
			expected: nil,
		},
		"deprecated-invalid": {
			input:    Number{Null: true, Unknown: true},
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
