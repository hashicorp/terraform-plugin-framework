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

func TestInt64TypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: NewInt64Value(123),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: NewInt64Unknown(),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: NewInt64Null(),
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

			got, err := Int64Type{}.ValueFromTerraform(ctx, test.input)
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

func TestInt64ValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int64Value
		expectation interface{}
	}
	tests := map[string]testCase{
		"known": {
			input:       NewInt64Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
		},
		"unknown": {
			input:       NewInt64Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       NewInt64Null(),
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

func TestInt64ValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int64Value
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       NewInt64Value(123),
			candidate:   NewInt64Value(123),
			expectation: true,
		},
		"known-known-diff": {
			input:       NewInt64Value(123),
			candidate:   NewInt64Value(456),
			expectation: false,
		},
		"known-unknown": {
			input:       NewInt64Value(123),
			candidate:   NewInt64Unknown(),
			expectation: false,
		},
		"known-null": {
			input:       NewInt64Value(123),
			candidate:   NewInt64Null(),
			expectation: false,
		},
		"unknown-value": {
			input:       NewInt64Unknown(),
			candidate:   NewInt64Value(123),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NewInt64Unknown(),
			candidate:   NewInt64Unknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       NewInt64Unknown(),
			candidate:   NewInt64Null(),
			expectation: false,
		},
		"null-known": {
			input:       NewInt64Null(),
			candidate:   NewInt64Value(123),
			expectation: false,
		},
		"null-unknown": {
			input:       NewInt64Null(),
			candidate:   NewInt64Unknown(),
			expectation: false,
		},
		"null-null": {
			input:       NewInt64Null(),
			candidate:   NewInt64Null(),
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

func TestInt64ValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int64Value
		expected bool
	}{
		"known": {
			input:    NewInt64Value(24),
			expected: false,
		},
		"null": {
			input:    NewInt64Null(),
			expected: true,
		},
		"unknown": {
			input:    NewInt64Unknown(),
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

func TestInt64ValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int64Value
		expected bool
	}{
		"known": {
			input:    NewInt64Value(24),
			expected: false,
		},
		"null": {
			input:    NewInt64Null(),
			expected: false,
		},
		"unknown": {
			input:    NewInt64Unknown(),
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

func TestInt64ValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int64Value
		expectation string
	}
	tests := map[string]testCase{
		"known-less-than-one": {
			input:       NewInt64Value(-12340984302980000),
			expectation: "-12340984302980000",
		},
		"known-more-than-one": {
			input:       NewInt64Value(92387938173219327),
			expectation: "92387938173219327",
		},
		"known-min-int64": {
			input:       NewInt64Value(math.MinInt64),
			expectation: "-9223372036854775808",
		},
		"known-max-int64": {
			input:       NewInt64Value(math.MaxInt64),
			expectation: "9223372036854775807",
		},
		"unknown": {
			input:       NewInt64Unknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewInt64Null(),
			expectation: "<null>",
		},
		"zero-value": {
			input:       Int64Value{},
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

func TestInt64ValueValueInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int64Value
		expected int64
	}{
		"known": {
			input:    NewInt64Value(24),
			expected: 24,
		},
		"null": {
			input:    NewInt64Null(),
			expected: 0,
		},
		"unknown": {
			input:    NewInt64Unknown(),
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
