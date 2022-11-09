package types

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

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
			expectation: StringValue("hello"),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectation: StringUnknown(),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.String, nil),
			expectation: StringNull(),
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
		"null": {
			input:    StringNull(),
			expected: true,
		},
		"unknown": {
			input:    StringUnknown(),
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
		"null": {
			input:    StringNull(),
			expected: false,
		},
		"unknown": {
			input:    StringUnknown(),
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
		"zero-value": {
			input:       String{},
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
		"null": {
			input:    StringNull(),
			expected: "",
		},
		"unknown": {
			input:    StringUnknown(),
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
