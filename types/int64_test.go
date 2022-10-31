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

func TestInt64ValueFromTerraform(t *testing.T) {
	t.Parallel()

	testInt64ValueFromTerraform(t, true)
}

func testInt64ValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: Int64Value(123),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: Int64Unknown(),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: Int64Null(),
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

			f := Int64Type.ValueFromTerraform
			if direct {
				f = int64ValueFromTerraform
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

func TestInt64ToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int64
		expectation interface{}
	}
	tests := map[string]testCase{
		"known": {
			input:       Int64Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
		},
		"unknown": {
			input:       Int64Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       Int64Null(),
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

func TestInt64Equal(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int64
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       Int64Value(123),
			candidate:   Int64Value(123),
			expectation: true,
		},
		"known-known-diff": {
			input:       Int64Value(123),
			candidate:   Int64Value(456),
			expectation: false,
		},
		"known-unknown": {
			input:       Int64Value(123),
			candidate:   Int64Unknown(),
			expectation: false,
		},
		"known-null": {
			input:       Int64Value(123),
			candidate:   Int64Null(),
			expectation: false,
		},
		"unknown-value": {
			input:       Int64Unknown(),
			candidate:   Int64Value(123),
			expectation: false,
		},
		"unknown-unknown": {
			input:       Int64Unknown(),
			candidate:   Int64Unknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       Int64Unknown(),
			candidate:   Int64Null(),
			expectation: false,
		},
		"null-known": {
			input:       Int64Null(),
			candidate:   Int64Value(123),
			expectation: false,
		},
		"null-unknown": {
			input:       Int64Null(),
			candidate:   Int64Unknown(),
			expectation: false,
		},
		"null-null": {
			input:       Int64Null(),
			candidate:   Int64Null(),
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

func TestInt64IsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int64
		expected bool
	}{
		"known": {
			input:    Int64Value(24),
			expected: false,
		},
		"null": {
			input:    Int64Null(),
			expected: true,
		},
		"unknown": {
			input:    Int64Unknown(),
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

func TestInt64IsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int64
		expected bool
	}{
		"known": {
			input:    Int64Value(24),
			expected: false,
		},
		"null": {
			input:    Int64Null(),
			expected: false,
		},
		"unknown": {
			input:    Int64Unknown(),
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

func TestInt64String(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int64
		expectation string
	}
	tests := map[string]testCase{
		"known-less-than-one": {
			input:       Int64Value(-12340984302980000),
			expectation: "-12340984302980000",
		},
		"known-more-than-one": {
			input:       Int64Value(92387938173219327),
			expectation: "92387938173219327",
		},
		"known-min-int64": {
			input:       Int64Value(math.MinInt64),
			expectation: "-9223372036854775808",
		},
		"known-max-int64": {
			input:       Int64Value(math.MaxInt64),
			expectation: "9223372036854775807",
		},
		"unknown": {
			input:       Int64Unknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       Int64Null(),
			expectation: "<null>",
		},
		"zero-value": {
			input:       Int64{},
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

func TestInt64ValueInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int64
		expected int64
	}{
		"known": {
			input:    Int64Value(24),
			expected: 24,
		},
		"null": {
			input:    Int64Null(),
			expected: 0,
		},
		"unknown": {
			input:    Int64Unknown(),
			expected: 0,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueInt64()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
