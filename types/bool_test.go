package types

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

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
			expectation: BoolValue(true),
		},
		"false": {
			input:       tftypes.NewValue(tftypes.Bool, false),
			expectation: BoolValue(false),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			expectation: BoolUnknown(),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Bool, nil),
			expectation: BoolNull(),
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
			candidate:   StringValue("true"),
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
		"known-true-null": {
			input:       BoolValue(true),
			candidate:   BoolNull(),
			expectation: false,
		},
		"known-true-unknown": {
			input:       BoolValue(true),
			candidate:   BoolUnknown(),
			expectation: false,
		},
		"known-false-nil": {
			input:       BoolValue(false),
			candidate:   nil,
			expectation: false,
		},
		"known-false-wrongtype": {
			input:       BoolValue(false),
			candidate:   StringValue("false"),
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
		"known-false-null": {
			input:       BoolValue(false),
			candidate:   BoolNull(),
			expectation: false,
		},
		"known-false-unknown": {
			input:       BoolValue(false),
			candidate:   BoolUnknown(),
			expectation: false,
		},
		"null-nil": {
			input:       BoolNull(),
			candidate:   nil,
			expectation: false,
		},
		"null-wrongtype": {
			input:       BoolNull(),
			candidate:   StringValue("true"),
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
		"null-null": {
			input:       BoolNull(),
			candidate:   BoolNull(),
			expectation: true,
		},
		"null-unknown": {
			input:       BoolNull(),
			candidate:   BoolUnknown(),
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
		"null": {
			input:    BoolNull(),
			expected: true,
		},
		"unknown": {
			input:    BoolUnknown(),
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
		"null": {
			input:    BoolNull(),
			expected: false,
		},
		"unknown": {
			input:    BoolUnknown(),
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
		"null": {
			input:    BoolNull(),
			expected: false,
		},
		"unknown": {
			input:    BoolUnknown(),
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
