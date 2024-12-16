// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/refinement"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
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
		"unknown-with-notnull-refinement": {
			input: NewBoolUnknown().RefineAsNotNull(),
			expectation: tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness: tfrefinement.NewNullness(false),
			}),
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
		"unknown-unknown-with-notnull-refinement": {
			input:       NewBoolUnknown(),
			candidate:   NewBoolUnknown().RefineAsNotNull(),
			expectation: false,
		},
		"unknowns-with-matching-notnull-refinements": {
			input:       NewBoolUnknown().RefineAsNotNull(),
			candidate:   NewBoolUnknown().RefineAsNotNull(),
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
		"unknown-with-notnull-refinement": {
			input:       NewBoolUnknown().RefineAsNotNull(),
			expectation: "<unknown, not null>",
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

func TestBoolValue_NotNullRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           BoolValue
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewBoolValue(true).RefineAsNotNull(),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewBoolNull().RefineAsNotNull(),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewBoolUnknown(),
			expectedFound: false,
		},
		"unknown-with-notnull-refinement": {
			input:           NewBoolUnknown().RefineAsNotNull(),
			expectedRefnVal: refinement.NewNotNull(),
			expectedFound:   true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, found := test.input.NotNullRefinement()
			if found != test.expectedFound {
				t.Fatalf("Expected refinement exists to be: %t, got: %t", test.expectedFound, found)
			}

			if got == nil && test.expectedRefnVal == nil {
				// Success!
				return
			}

			if got == nil && test.expectedRefnVal != nil {
				t.Fatalf("Expected refinement data: <%+v>, got: nil", test.expectedRefnVal)
			}

			if diff := cmp.Diff(*got, test.expectedRefnVal); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
