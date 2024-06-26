// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func TestInt32ValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int32Value
		expectation interface{}
	}
	tests := map[string]testCase{
		"known": {
			input:       NewInt32Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
		},
		"unknown": {
			input:       NewInt32Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       NewInt32Null(),
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

func TestInt32ValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int32Value
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       NewInt32Value(123),
			candidate:   NewInt32Value(123),
			expectation: true,
		},
		"known-known-diff": {
			input:       NewInt32Value(123),
			candidate:   NewInt32Value(456),
			expectation: false,
		},
		"known-unknown": {
			input:       NewInt32Value(123),
			candidate:   NewInt32Unknown(),
			expectation: false,
		},
		"known-null": {
			input:       NewInt32Value(123),
			candidate:   NewInt32Null(),
			expectation: false,
		},
		"unknown-value": {
			input:       NewInt32Unknown(),
			candidate:   NewInt32Value(123),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NewInt32Unknown(),
			candidate:   NewInt32Unknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       NewInt32Unknown(),
			candidate:   NewInt32Null(),
			expectation: false,
		},
		"null-known": {
			input:       NewInt32Null(),
			candidate:   NewInt32Value(123),
			expectation: false,
		},
		"null-unknown": {
			input:       NewInt32Null(),
			candidate:   NewInt32Unknown(),
			expectation: false,
		},
		"null-null": {
			input:       NewInt32Null(),
			candidate:   NewInt32Null(),
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

func TestInt32ValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int32Value
		expected bool
	}{
		"known": {
			input:    NewInt32Value(24),
			expected: false,
		},
		"null": {
			input:    NewInt32Null(),
			expected: true,
		},
		"unknown": {
			input:    NewInt32Unknown(),
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

func TestInt32ValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int32Value
		expected bool
	}{
		"known": {
			input:    NewInt32Value(24),
			expected: false,
		},
		"null": {
			input:    NewInt32Null(),
			expected: false,
		},
		"unknown": {
			input:    NewInt32Unknown(),
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

func TestInt32ValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int32Value
		expectation string
	}
	tests := map[string]testCase{
		"known-less-than-one": {
			input:       NewInt32Value(-2147183641),
			expectation: "-2147183641",
		},
		"known-more-than-one": {
			input:       NewInt32Value(2147483620),
			expectation: "2147483620",
		},
		"known-min-int32": {
			input:       NewInt32Value(math.MinInt32),
			expectation: "-2147483648",
		},
		"known-max-int32": {
			input:       NewInt32Value(math.MaxInt32),
			expectation: "2147483647",
		},
		"unknown": {
			input:       NewInt32Unknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewInt32Null(),
			expectation: "<null>",
		},
		"zero-value": {
			input:       Int32Value{},
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

func TestInt32ValueValueInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int32Value
		expected int32
	}{
		"known": {
			input:    NewInt32Value(24),
			expected: 24,
		},
		"null": {
			input:    NewInt32Null(),
			expected: 0,
		},
		"unknown": {
			input:    NewInt32Unknown(),
			expected: 0,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueInt32()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestInt32ValueValueInt32Pointer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Int32Value
		expected *int32
	}{
		"known": {
			input:    NewInt32Value(24),
			expected: pointer(int32(24)),
		},
		"null": {
			input:    NewInt32Null(),
			expected: nil,
		},
		"unknown": {
			input:    NewInt32Unknown(),
			expected: pointer(int32(0)),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueInt32Pointer()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNewInt32PointerValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value    *int32
		expected Int32Value
	}{
		"nil": {
			value:    nil,
			expected: NewInt32Null(),
		},
		"value": {
			value:    pointer(int32(123)),
			expected: NewInt32Value(123),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := NewInt32PointerValue(testCase.value)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
