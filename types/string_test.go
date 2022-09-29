package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestStringValueDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	knownString := StringValue("test")

	knownString.Null = true

	if knownString.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	knownString.Unknown = true

	if knownString.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	knownString.Value = "not-test"

	if knownString.ValueString() == "not-test" {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestStringNullDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	nullString := StringNull()

	nullString.Null = false

	if !nullString.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	nullString.Unknown = true

	if nullString.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	nullString.Value = "test"

	if nullString.ValueString() == "test" {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestStringUnknownDeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	unknownString := StringUnknown()

	unknownString.Null = true

	if unknownString.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	unknownString.Unknown = false

	if !unknownString.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	unknownString.Value = "test"

	if unknownString.ValueString() == "test" {
		t.Error("unexpected value update after Value field setting")
	}
}

func TestStringValueFromTerraform(t *testing.T) {
	t.Parallel()

	testStringValueFromTerraform(t, true)
}

func testStringValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"true": {
			input:       tftypes.NewValue(tftypes.String, "hello"),
			expectation: String{Value: "hello"},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectation: String{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.String, nil),
			expectation: String{Null: true},
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectedErr: "can't unmarshal tftypes.Number into *string, expected string",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			f := StringType.ValueFromTerraform
			if direct {
				f = stringValueFromTerraform
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

func TestStringToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       String
		expectation interface{}
	}
	tests := map[string]testCase{
		"known": {
			input:       StringValue("test"),
			expectation: tftypes.NewValue(tftypes.String, "test"),
		},
		"unknown": {
			input:       StringUnknown(),
			expectation: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"null": {
			input:       StringNull(),
			expectation: tftypes.NewValue(tftypes.String, nil),
		},
		"deprecated-known": {
			input:       String{Value: "hello"},
			expectation: tftypes.NewValue(tftypes.String, "hello"),
		},
		"deprecated-unknown": {
			input:       String{Unknown: true},
			expectation: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       String{Null: true},
			expectation: tftypes.NewValue(tftypes.String, nil),
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

func TestStringEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       String
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       StringValue("test"),
			candidate:   StringValue("test"),
			expectation: true,
		},
		"known-known-diff": {
			input:       StringValue("test"),
			candidate:   StringValue("not-test"),
			expectation: false,
		},
		"known-unknown": {
			input:       StringValue("test"),
			candidate:   StringUnknown(),
			expectation: false,
		},
		"known-null": {
			input:       StringValue("test"),
			candidate:   StringNull(),
			expectation: false,
		},
		"unknown-value": {
			input:       StringUnknown(),
			candidate:   StringValue("test"),
			expectation: false,
		},
		"unknown-unknown": {
			input:       StringUnknown(),
			candidate:   StringUnknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       StringUnknown(),
			candidate:   StringNull(),
			expectation: false,
		},
		"null-known": {
			input:       StringNull(),
			candidate:   StringValue("test"),
			expectation: false,
		},
		"null-unknown": {
			input:       StringNull(),
			candidate:   StringUnknown(),
			expectation: false,
		},
		"null-null": {
			input:       StringNull(),
			candidate:   StringNull(),
			expectation: true,
		},
		"deprecated-known-known-same": {
			input:       String{Value: "test"},
			candidate:   StringValue("test"),
			expectation: false, // intentional
		},
		"deprecated-known-known-diff": {
			input:       String{Value: "test"},
			candidate:   StringValue("not-test"),
			expectation: false,
		},
		"deprecated-known-unknown": {
			input:       String{Value: "test"},
			candidate:   StringUnknown(),
			expectation: false,
		},
		"deprecated-known-null": {
			input:       String{Value: "test"},
			candidate:   StringNull(),
			expectation: false,
		},
		"deprecated-known-deprecated-known-same": {
			input:       String{Value: "hello"},
			candidate:   String{Value: "hello"},
			expectation: true,
		},
		"deprecated-known-deprecated-known-diff": {
			input:       String{Value: "hello"},
			candidate:   String{Value: "world"},
			expectation: false,
		},
		"deprecated-known-deprecated-unknown": {
			input:       String{Value: "hello"},
			candidate:   String{Unknown: true},
			expectation: false,
		},
		"deprecated-known-deprecated-null": {
			input:       String{Value: "hello"},
			candidate:   String{Null: true},
			expectation: false,
		},
		"deprecated-known-wrongType": {
			input:       String{Value: "hello"},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"deprecated-known-nil": {
			input:       String{Value: "hello"},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-unknown-value": {
			input:       String{Unknown: true},
			candidate:   StringValue("test"),
			expectation: false,
		},
		"deprecated-unknown-unknown": {
			input:       String{Unknown: true},
			candidate:   StringUnknown(),
			expectation: false, // intentional
		},
		"deprecated-unknown-null": {
			input:       String{Unknown: true},
			candidate:   StringNull(),
			expectation: false,
		},
		"deprecated-unknown-deprecated-known": {
			input:       String{Unknown: true},
			candidate:   String{Value: "hello"},
			expectation: false,
		},
		"deprecated-unknown-deprecated-unknown": {
			input:       String{Unknown: true},
			candidate:   String{Unknown: true},
			expectation: true,
		},
		"deprecated-unknown-deprecated-null": {
			input:       String{Unknown: true},
			candidate:   String{Null: true},
			expectation: false,
		},
		"deprecated-unknown-wrongType": {
			input:       String{Unknown: true},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"deprecated-unknown-nil": {
			input:       String{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"deprecated-null-known": {
			input:       String{Null: true},
			candidate:   StringValue("test"),
			expectation: false,
		},
		"deprecated-null-unknown": {
			input:       String{Null: true},
			candidate:   StringUnknown(),
			expectation: false,
		},
		"deprecated-null-null": {
			input:       String{Null: true},
			candidate:   StringNull(),
			expectation: false, // intentional
		},
		"deprecated-null-deprecated-known": {
			input:       String{Null: true},
			candidate:   String{Value: "hello"},
			expectation: false,
		},
		"deprecated-null-deprecated-unknown": {
			input:       String{Null: true},
			candidate:   String{Unknown: true},
			expectation: false,
		},
		"deprecated-null-deprecated-null": {
			input:       String{Null: true},
			candidate:   String{Null: true},
			expectation: true,
		},
		"deprecated-null-wrongType": {
			input:       String{Null: true},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"deprecated-null-nil": {
			input:       String{Null: true},
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

func TestStringIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    String
		expected bool
	}{
		"known": {
			input:    StringValue("test"),
			expected: false,
		},
		"deprecated-known": {
			input:    String{Value: "test"},
			expected: false,
		},
		"null": {
			input:    StringNull(),
			expected: true,
		},
		"deprecated-null": {
			input:    String{Null: true},
			expected: true,
		},
		"unknown": {
			input:    StringUnknown(),
			expected: false,
		},
		"deprecated-unknown": {
			input:    String{Unknown: true},
			expected: false,
		},
		"deprecated-invalid": {
			input:    String{Null: true, Unknown: true},
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

func TestStringIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    String
		expected bool
	}{
		"known": {
			input:    StringValue("test"),
			expected: false,
		},
		"deprecated-known": {
			input:    String{Value: "test"},
			expected: false,
		},
		"null": {
			input:    StringNull(),
			expected: false,
		},
		"deprecated-null": {
			input:    String{Null: true},
			expected: false,
		},
		"unknown": {
			input:    StringUnknown(),
			expected: true,
		},
		"deprecated-unknown": {
			input:    String{Unknown: true},
			expected: true,
		},
		"deprecated-invalid": {
			input:    String{Null: true, Unknown: true},
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

func TestStringString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       String
		expectation string
	}
	tests := map[string]testCase{
		"known-non-empty": {
			input:       StringValue("test"),
			expectation: `"test"`,
		},
		"known-empty": {
			input:       StringValue(""),
			expectation: `""`,
		},
		"known-quotes": {
			input:       StringValue(`testing is "fun"`),
			expectation: `"testing is \"fun\""`,
		},
		"unknown": {
			input:       StringUnknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       StringNull(),
			expectation: "<null>",
		},
		"deprecated-known-non-empty": {
			input:       String{Value: "simple"},
			expectation: `"simple"`,
		},
		"deprecated-known-empty": {
			input:       String{Value: ""},
			expectation: `""`,
		},
		"deprecated-known-quotes": {
			input:       String{Value: `testing is "fun"`},
			expectation: `"testing is \"fun\""`,
		},
		"deprecated-unknown": {
			input:       String{Unknown: true},
			expectation: "<unknown>",
		},
		"deprecated-null": {
			input:       String{Null: true},
			expectation: "<null>",
		},
		"default-0": {
			input:       String{},
			expectation: `""`,
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

func TestStringValueString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    String
		expected string
	}{
		"known": {
			input:    StringValue("test"),
			expected: "test",
		},
		"deprecated-known": {
			input:    String{Value: "test"},
			expected: "test",
		},
		"null": {
			input:    StringNull(),
			expected: "",
		},
		"deprecated-null": {
			input:    String{Null: true},
			expected: "",
		},
		"unknown": {
			input:    StringUnknown(),
			expected: "",
		},
		"deprecated-unknown": {
			input:    String{Unknown: true},
			expected: "",
		},
		"deprecated-invalid": {
			input:    String{Null: true, Unknown: true},
			expected: "",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueString()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
