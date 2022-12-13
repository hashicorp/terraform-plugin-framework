package basetypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStringTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"true": {
			input:       tftypes.NewValue(tftypes.String, "hello"),
			expectation: NewStringValue("hello"),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectation: NewStringUnknown(),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.String, nil),
			expectation: NewStringNull(),
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

			got, err := StringType{}.ValueFromTerraform(ctx, test.input)
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

func TestStringValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       StringValue
		expectation interface{}
	}
	tests := map[string]testCase{
		"known": {
			input:       NewStringValue("test"),
			expectation: tftypes.NewValue(tftypes.String, "test"),
		},
		"unknown": {
			input:       NewStringUnknown(),
			expectation: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"null": {
			input:       NewStringNull(),
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

func TestStringValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       StringValue
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       NewStringValue("test"),
			candidate:   NewStringValue("test"),
			expectation: true,
		},
		"known-known-diff": {
			input:       NewStringValue("test"),
			candidate:   NewStringValue("not-test"),
			expectation: false,
		},
		"known-unknown": {
			input:       NewStringValue("test"),
			candidate:   NewStringUnknown(),
			expectation: false,
		},
		"known-null": {
			input:       NewStringValue("test"),
			candidate:   NewStringNull(),
			expectation: false,
		},
		"unknown-value": {
			input:       NewStringUnknown(),
			candidate:   NewStringValue("test"),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NewStringUnknown(),
			candidate:   NewStringUnknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       NewStringUnknown(),
			candidate:   NewStringNull(),
			expectation: false,
		},
		"null-known": {
			input:       NewStringNull(),
			candidate:   NewStringValue("test"),
			expectation: false,
		},
		"null-unknown": {
			input:       NewStringNull(),
			candidate:   NewStringUnknown(),
			expectation: false,
		},
		"null-null": {
			input:       NewStringNull(),
			candidate:   NewStringNull(),
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

func TestStringValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    StringValue
		expected bool
	}{
		"known": {
			input:    NewStringValue("test"),
			expected: false,
		},
		"null": {
			input:    NewStringNull(),
			expected: true,
		},
		"unknown": {
			input:    NewStringUnknown(),
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

func TestStringValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    StringValue
		expected bool
	}{
		"known": {
			input:    NewStringValue("test"),
			expected: false,
		},
		"null": {
			input:    NewStringNull(),
			expected: false,
		},
		"unknown": {
			input:    NewStringUnknown(),
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

func TestStringValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       StringValue
		expectation string
	}
	tests := map[string]testCase{
		"known-non-empty": {
			input:       NewStringValue("test"),
			expectation: `"test"`,
		},
		"known-empty": {
			input:       NewStringValue(""),
			expectation: `""`,
		},
		"known-quotes": {
			input:       NewStringValue(`testing is "fun"`),
			expectation: `"testing is \"fun\""`,
		},
		"unknown": {
			input:       NewStringUnknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewStringNull(),
			expectation: "<null>",
		},
		"zero-value": {
			input:       StringValue{},
			expectation: `<null>`,
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

func TestStringValueValueString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    StringValue
		expected string
	}{
		"known": {
			input:    NewStringValue("test"),
			expected: "test",
		},
		"null": {
			input:    NewStringNull(),
			expected: "",
		},
		"unknown": {
			input:    NewStringUnknown(),
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
