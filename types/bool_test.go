package types

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestBoolValueDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	knownBool := BoolValue(false)

	knownBool.Null = true

	if knownBool.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	knownBool.Unknown = true

	if knownBool.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	knownBool.Value = true

	if knownBool.ValueBool() {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestBoolNullDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	nullBool := BoolNull()

	nullBool.Null = false

	if !nullBool.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	nullBool.Unknown = true

	if nullBool.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	nullBool.Value = true

	if nullBool.ValueBool() {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestBoolUnknownDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	unknownBool := BoolUnknown()

	unknownBool.Null = true

	if unknownBool.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	unknownBool.Unknown = false

	if !unknownBool.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	unknownBool.Value = true

	if unknownBool.ValueBool() {
		t.Error("unexpected value update after Value field setting")
	}
}

func TestBoolValueFromTerraform(t *testing.T) {
	t.Parallel()

	testBoolValueFromTerraform(t, true)
}

func testBoolValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"true": {
			input:       tftypes.NewValue(tftypes.Bool, true),
			expectation: Bool{Value: true},
		},
		"false": {
			input:       tftypes.NewValue(tftypes.Bool, false),
			expectation: Bool{Value: false},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			expectation: Bool{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Bool, nil),
			expectation: Bool{Null: true},
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *bool, expected boolean",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			f := BoolType.ValueFromTerraform
			if direct {
				f = boolValueFromTerraform
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

func TestBoolToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Bool
		expectation interface{}
	}
	tests := map[string]testCase{
		"known-true": {
			input:       BoolValue(true),
			expectation: tftypes.NewValue(tftypes.Bool, true),
		},
		"known-false": {
			input:       BoolValue(false),
			expectation: tftypes.NewValue(tftypes.Bool, false),
		},
		"unknown": {
			input:       BoolUnknown(),
			expectation: tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
		},
		"null": {
			input:       BoolNull(),
			expectation: tftypes.NewValue(tftypes.Bool, nil),
		},
		"deprecated-true": {
			input:       Bool{Value: true},
			expectation: tftypes.NewValue(tftypes.Bool, true),
		},
		"deprecated-false": {
			input:       Bool{Value: false},
			expectation: tftypes.NewValue(tftypes.Bool, false),
		},
		"deprecated-unknown": {
			input:       Bool{Unknown: true},
			expectation: tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       Bool{Null: true},
			expectation: tftypes.NewValue(tftypes.Bool, nil),
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
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %+v, got %+v", test.expectation, got)
			}
		})
	}
}

func TestBoolEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Bool
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-true-nil": {
			input:       BoolValue(true),
			candidate:   nil,
			expectation: false,
		},
		"known-true-wrongtype": {
			input:       BoolValue(true),
			candidate:   String{Value: "true"},
			expectation: false,
		},
		"known-true-known-false": {
			input:       BoolValue(true),
			candidate:   BoolValue(false),
			expectation: false,
		},
		"known-true-known-true": {
			input:       BoolValue(true),
			candidate:   BoolValue(true),
			expectation: true,
		},
		"known-true-deprecated-false": {
			input:       BoolValue(true),
			candidate:   Bool{Value: false},
			expectation: false,
		},
		"known-true-deprecated-true": {
			input:       BoolValue(true),
			candidate:   Bool{Value: true},
			expectation: false, // intentional
		},
		"known-true-null": {
			input:       BoolValue(true),
			candidate:   BoolNull(),
			expectation: false,
		},
		"known-true-deprecated-null": {
			input:       BoolValue(true),
			candidate:   Bool{Null: true},
			expectation: false,
		},
		"known-true-unknown": {
			input:       BoolValue(true),
			candidate:   BoolUnknown(),
			expectation: false,
		},
		"known-true-deprecated-unknown": {
			input:       BoolValue(true),
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"known-false-nil": {
			input:       BoolValue(false),
			candidate:   nil,
			expectation: false,
		},
		"known-false-wrongtype": {
			input:       BoolValue(false),
			candidate:   String{Value: "false"},
			expectation: false,
		},
		"known-false-known-false": {
			input:       BoolValue(false),
			candidate:   BoolValue(false),
			expectation: true,
		},
		"known-false-known-true": {
			input:       BoolValue(false),
			candidate:   BoolValue(true),
			expectation: false,
		},
		"known-false-deprecated-false": {
			input:       BoolValue(false),
			candidate:   Bool{Value: false},
			expectation: false, // intentional
		},
		"known-false-deprecated-true": {
			input:       BoolValue(false),
			candidate:   Bool{Value: true},
			expectation: false,
		},
		"known-false-null": {
			input:       BoolValue(false),
			candidate:   BoolNull(),
			expectation: false,
		},
		"known-false-deprecated-null": {
			input:       BoolValue(false),
			candidate:   Bool{Null: true},
			expectation: false,
		},
		"known-false-unknown": {
			input:       BoolValue(false),
			candidate:   BoolUnknown(),
			expectation: false,
		},
		"known-false-deprecated-unknown": {
			input:       BoolValue(false),
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"null-nil": {
			input:       BoolNull(),
			candidate:   nil,
			expectation: false,
		},
		"null-wrongtype": {
			input:       BoolNull(),
			candidate:   String{Value: "true"},
			expectation: false,
		},
		"null-known-false": {
			input:       BoolNull(),
			candidate:   BoolValue(false),
			expectation: false,
		},
		"null-known-true": {
			input:       BoolNull(),
			candidate:   BoolValue(true),
			expectation: false,
		},
		"null-deprecated-true": {
			input:       BoolNull(),
			candidate:   Bool{Value: true},
			expectation: false,
		},
		"null-deprecated-false": {
			input:       BoolNull(),
			candidate:   Bool{Value: false},
			expectation: false,
		},
		"null-null": {
			input:       BoolNull(),
			candidate:   BoolNull(),
			expectation: true,
		},
		"null-deprecated-null": {
			input:       BoolNull(),
			candidate:   Bool{Null: true},
			expectation: false, // intentional
		},
		"null-unknown": {
			input:       BoolNull(),
			candidate:   BoolUnknown(),
			expectation: false,
		},
		"null-deprecated-unknown": {
			input:       BoolNull(),
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"deprecated-true-known-false": {
			input:       Bool{Value: true},
			candidate:   BoolValue(false),
			expectation: false,
		},
		"deprecated-true-known-true": {
			input:       Bool{Value: true},
			candidate:   BoolValue(true),
			expectation: false, // intentional
		},
		"deprecated-true-deprecated-true": {
			input:       Bool{Value: true},
			candidate:   Bool{Value: true},
			expectation: true,
		},
		"deprecated-true-deprecated-false": {
			input:       Bool{Value: true},
			candidate:   Bool{Value: false},
			expectation: false,
		},
		"deprecated-true-unknown": {
			input:       Bool{Value: true},
			candidate:   BoolUnknown(),
			expectation: false,
		},
		"deprecated-true-deprecated-unknown": {
			input:       Bool{Value: true},
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"deprecated-true-null": {
			input:       Bool{Value: true},
			candidate:   BoolNull(),
			expectation: false,
		},
		"deprecated-true-deprecated-null": {
			input:       Bool{Value: true},
			candidate:   Bool{Null: true},
			expectation: false,
		},
		"deprecated-true-wrongType": {
			input:       Bool{Value: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-true-nil": {
			input:       Bool{Value: true},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-false-known-false": {
			input:       Bool{Value: false},
			candidate:   BoolValue(false),
			expectation: false, // intentional
		},
		"deprecated-false-known-true": {
			input:       Bool{Value: false},
			candidate:   BoolValue(true),
			expectation: false,
		},
		"deprecated-false-deprecated-true": {
			input:       Bool{Value: false},
			candidate:   Bool{Value: true},
			expectation: false,
		},
		"deprecated-false-deprecated-false": {
			input:       Bool{Value: false},
			candidate:   Bool{Value: false},
			expectation: true,
		},
		"deprecated-false-unknown": {
			input:       Bool{Value: false},
			candidate:   BoolUnknown(),
			expectation: false,
		},
		"deprecated-false-deprecated-unknown": {
			input:       Bool{Value: false},
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"deprecated-false-null": {
			input:       Bool{Value: false},
			candidate:   BoolNull(),
			expectation: false,
		},
		"deprecated-false-deprecated-null": {
			input:       Bool{Value: false},
			candidate:   Bool{Null: true},
			expectation: false,
		},
		"deprecated-false-wrongType": {
			input:       Bool{Value: false},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-false-nil": {
			input:       Bool{Value: false},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-unknown-known-false": {
			input:       Bool{Unknown: true},
			candidate:   BoolValue(false),
			expectation: false,
		},
		"deprecated-unknown-known-true": {
			input:       Bool{Unknown: true},
			candidate:   BoolValue(true),
			expectation: false,
		},
		"deprecated-unknown-deprecated-true": {
			input:       Bool{Unknown: true},
			candidate:   Bool{Value: true},
			expectation: false,
		},
		"deprecated-unknown-deprecated-false": {
			input:       Bool{Unknown: true},
			candidate:   Bool{Value: false},
			expectation: false,
		},
		"deprecated-unknown-deprecated-unknown": {
			input:       Bool{Unknown: true},
			candidate:   Bool{Unknown: true},
			expectation: true,
		},
		"deprecated-unknown-deprecated-null": {
			input:       Bool{Unknown: true},
			candidate:   Bool{Null: true},
			expectation: false,
		},
		"deprecated-unknown-wrongType": {
			input:       Bool{Unknown: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-unknown-nil": {
			input:       Bool{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-null-deprecated-true": {
			input:       Bool{Null: true},
			candidate:   Bool{Value: true},
			expectation: false,
		},
		"deprecated-null-deprecated-false": {
			input:       Bool{Null: true},
			candidate:   Bool{Value: false},
			expectation: false,
		},
		"deprecated-null-known-false": {
			input:       Bool{Null: true},
			candidate:   BoolValue(false),
			expectation: false,
		},
		"deprecated-null-known-true": {
			input:       Bool{Null: true},
			candidate:   BoolValue(true),
			expectation: false,
		},
		"deprecated-null-unknown": {
			input:       Bool{Null: true},
			candidate:   BoolUnknown(),
			expectation: false,
		},
		"deprecated-null-deprecated-unknown": {
			input:       Bool{Null: true},
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"deprecated-null-null": {
			input:       Bool{Null: true},
			candidate:   BoolNull(),
			expectation: false, // intentional
		},
		"deprecated-null-deprecated-null": {
			input:       Bool{Null: true},
			candidate:   Bool{Null: true},
			expectation: true,
		},
		"deprecated-null-wrongType": {
			input:       Bool{Null: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"deprecated-null-nil": {
			input:       Bool{Null: true},
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

func TestBoolIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Bool
		expected bool
	}{
		"known": {
			input:    BoolValue(true),
			expected: false,
		},
		"deprecated-known": {
			input:    Bool{Value: true},
			expected: false,
		},
		"null": {
			input:    BoolNull(),
			expected: true,
		},
		"deprecated-null": {
			input:    Bool{Null: true},
			expected: true,
		},
		"unknown": {
			input:    BoolUnknown(),
			expected: false,
		},
		"deprecated-unknown": {
			input:    Bool{Unknown: true},
			expected: false,
		},
		"deprecated-invalid": {
			input:    Bool{Null: true, Unknown: true},
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

func TestBoolIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Bool
		expected bool
	}{
		"known": {
			input:    BoolValue(true),
			expected: false,
		},
		"deprecated-known": {
			input:    Bool{Value: true},
			expected: false,
		},
		"null": {
			input:    BoolNull(),
			expected: false,
		},
		"deprecated-null": {
			input:    Bool{Null: true},
			expected: false,
		},
		"unknown": {
			input:    BoolUnknown(),
			expected: true,
		},
		"deprecated-unknown": {
			input:    Bool{Unknown: true},
			expected: true,
		},
		"deprecated-invalid": {
			input:    Bool{Null: true, Unknown: true},
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

func TestBoolString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Bool
		expectation string
	}
	tests := map[string]testCase{
		"known-true": {
			input:       BoolValue(true),
			expectation: "true",
		},
		"known-false": {
			input:       BoolValue(false),
			expectation: "false",
		},
		"null": {
			input:       BoolNull(),
			expectation: "<null>",
		},
		"unknown": {
			input:       BoolUnknown(),
			expectation: "<unknown>",
		},
		"deprecated-true": {
			input:       Bool{Value: true},
			expectation: "true",
		},
		"deprecated-false": {
			input:       Bool{Value: false},
			expectation: "false",
		},
		"deprecated-unknown": {
			input:       Bool{Unknown: true},
			expectation: "<unknown>",
		},
		"deprecated-null": {
			input:       Bool{Null: true},
			expectation: "<null>",
		},
		"default-false": {
			input:       Bool{},
			expectation: "false",
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

func TestBoolValueBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Bool
		expected bool
	}{
		"known-false": {
			input:    BoolValue(false),
			expected: false,
		},
		"known-true": {
			input:    BoolValue(true),
			expected: true,
		},
		"deprecated-known-false": {
			input:    Bool{Value: false},
			expected: false,
		},
		"deprecated-known-true": {
			input:    Bool{Value: true},
			expected: true,
		},
		"null": {
			input:    BoolNull(),
			expected: false,
		},
		"deprecated-null": {
			input:    Bool{Null: true},
			expected: false,
		},
		"unknown": {
			input:    BoolUnknown(),
			expected: false,
		},
		"deprecated-unknown": {
			input:    Bool{Unknown: true},
			expected: false,
		},
		"deprecated-invalid": {
			input:    Bool{Null: true, Unknown: true},
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueBool()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
