// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBoolValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       BoolValue
		expectation interface{}
	}
	tests := map[string]testCase{
		"known-true": {
			input:       NewBoolValue(true),
			expectation: tftypes.NewValue(tftypes.Bool, true),
		},
		"known-false": {
			input:       NewBoolValue(false),
			expectation: tftypes.NewValue(tftypes.Bool, false),
		},
		"unknown": {
			input:       NewBoolUnknown(),
			expectation: tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
		},
		"null": {
			input:       NewBoolNull(),
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

func TestBoolValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       BoolValue
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-true-nil": {
			input:       NewBoolValue(true),
			candidate:   nil,
			expectation: false,
		},
		"known-true-wrongtype": {
			input:       NewBoolValue(true),
			candidate:   NewStringValue("true"),
			expectation: false,
		},
		"known-true-known-false": {
			input:       NewBoolValue(true),
			candidate:   NewBoolValue(false),
			expectation: false,
		},
		"known-true-known-true": {
			input:       NewBoolValue(true),
			candidate:   NewBoolValue(true),
			expectation: true,
		},
		"known-true-null": {
			input:       NewBoolValue(true),
			candidate:   NewBoolNull(),
			expectation: false,
		},
		"known-true-unknown": {
			input:       NewBoolValue(true),
			candidate:   NewBoolUnknown(),
			expectation: false,
		},
		"known-false-nil": {
			input:       NewBoolValue(false),
			candidate:   nil,
			expectation: false,
		},
		"known-false-wrongtype": {
			input:       NewBoolValue(false),
			candidate:   NewStringValue("false"),
			expectation: false,
		},
		"known-false-known-false": {
			input:       NewBoolValue(false),
			candidate:   NewBoolValue(false),
			expectation: true,
		},
		"known-false-known-true": {
			input:       NewBoolValue(false),
			candidate:   NewBoolValue(true),
			expectation: false,
		},
		"known-false-null": {
			input:       NewBoolValue(false),
			candidate:   NewBoolNull(),
			expectation: false,
		},
		"known-false-unknown": {
			input:       NewBoolValue(false),
			candidate:   NewBoolUnknown(),
			expectation: false,
		},
		"null-nil": {
			input:       NewBoolNull(),
			candidate:   nil,
			expectation: false,
		},
		"null-wrongtype": {
			input:       NewBoolNull(),
			candidate:   NewStringValue("true"),
			expectation: false,
		},
		"null-known-false": {
			input:       NewBoolNull(),
			candidate:   NewBoolValue(false),
			expectation: false,
		},
		"null-known-true": {
			input:       NewBoolNull(),
			candidate:   NewBoolValue(true),
			expectation: false,
		},
		"null-null": {
			input:       NewBoolNull(),
			candidate:   NewBoolNull(),
			expectation: true,
		},
		"null-unknown": {
			input:       NewBoolNull(),
			candidate:   NewBoolUnknown(),
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

func TestBoolValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    BoolValue
		expected bool
	}{
		"known": {
			input:    NewBoolValue(true),
			expected: false,
		},
		"null": {
			input:    NewBoolNull(),
			expected: true,
		},
		"unknown": {
			input:    NewBoolUnknown(),
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

func TestBoolValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    BoolValue
		expected bool
	}{
		"known": {
			input:    NewBoolValue(true),
			expected: false,
		},
		"null": {
			input:    NewBoolNull(),
			expected: false,
		},
		"unknown": {
			input:    NewBoolUnknown(),
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

func TestBoolValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       BoolValue
		expectation string
	}
	tests := map[string]testCase{
		"known-true": {
			input:       NewBoolValue(true),
			expectation: "true",
		},
		"known-false": {
			input:       NewBoolValue(false),
			expectation: "false",
		},
		"null": {
			input:       NewBoolNull(),
			expectation: "<null>",
		},
		"unknown": {
			input:       NewBoolUnknown(),
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

func TestBoolValueValueBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    BoolValue
		expected bool
	}{
		"known-false": {
			input:    NewBoolValue(false),
			expected: false,
		},
		"known-true": {
			input:    NewBoolValue(true),
			expected: true,
		},
		"null": {
			input:    NewBoolNull(),
			expected: false,
		},
		"unknown": {
			input:    NewBoolUnknown(),
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

func TestBoolValueValueBoolPointer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    BoolValue
		expected *bool
	}{
		"known-false": {
			input:    NewBoolValue(false),
			expected: pointer(false),
		},
		"known-true": {
			input:    NewBoolValue(true),
			expected: pointer(true),
		},
		"null": {
			input:    NewBoolNull(),
			expected: nil,
		},
		"unknown": {
			input:    NewBoolUnknown(),
			expected: pointer(false),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueBoolPointer()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNewBoolPointerValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value    *bool
		expected BoolValue
	}{
		"nil": {
			value:    nil,
			expected: NewBoolNull(),
		},
		"value": {
			value:    pointer(true),
			expected: NewBoolValue(true),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := NewBoolPointerValue(testCase.value)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
