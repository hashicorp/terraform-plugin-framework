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
		"unknown-with-notnull-refinement": {
			input: NewStringUnknown().RefineAsNotNull(),
			expectation: tftypes.NewValue(tftypes.String, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness: tfrefinement.NewNullness(false),
			}),
		},
		"unknown-with-prefix-refinement": {
			input: NewStringUnknown().RefineWithPrefix("hello://"),
			expectation: tftypes.NewValue(tftypes.String, tftypes.UnknownValue).Refine(tfrefinement.Refinements{
				tfrefinement.KeyNullness:     tfrefinement.NewNullness(false),
				tfrefinement.KeyStringPrefix: tfrefinement.NewStringPrefix("hello://"),
			}),
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
		"unknown-unknown-with-notnull-refinement": {
			input:       NewStringUnknown(),
			candidate:   NewStringUnknown().RefineAsNotNull(),
			expectation: false,
		},
		"unknown-unknown-with-prefix-refinement": {
			input:       NewStringUnknown(),
			candidate:   NewStringUnknown().RefineWithPrefix("hello://"),
			expectation: false,
		},
		"unknowns-with-matching-notnull-refinements": {
			input:       NewStringUnknown().RefineAsNotNull(),
			candidate:   NewStringUnknown().RefineAsNotNull(),
			expectation: true,
		},
		"unknowns-with-matching-prefix-refinements": {
			input:       NewStringUnknown().RefineWithPrefix("hello://"),
			candidate:   NewStringUnknown().RefineWithPrefix("hello://"),
			expectation: true,
		},
		"unknowns-with-different-prefix-refinements": {
			input:       NewStringUnknown().RefineWithPrefix("hello://"),
			candidate:   NewStringUnknown().RefineWithPrefix("world://"),
			expectation: false,
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
		"unknown-with-notnull-refinement": {
			input:       NewStringUnknown().RefineAsNotNull(),
			expectation: "<unknown, not null>",
		},
		"unknown-with-prefix-refinement": {
			input:       NewStringUnknown().RefineWithPrefix("hello://"),
			expectation: `<unknown, not null, prefix = "hello://">`,
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

func TestStringValueValueStringPointer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    StringValue
		expected *string
	}{
		"known": {
			input:    NewStringValue("test"),
			expected: pointer("test"),
		},
		"null": {
			input:    NewStringNull(),
			expected: nil,
		},
		"unknown": {
			input:    NewStringUnknown(),
			expected: pointer(""),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueStringPointer()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNewStringPointerValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value    *string
		expected StringValue
	}{
		"nil": {
			value:    nil,
			expected: NewStringNull(),
		},
		"value": {
			value:    pointer("test"),
			expected: NewStringValue("test"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := NewStringPointerValue(testCase.value)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestStringValue_NotNullRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           StringValue
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewStringValue("test").RefineAsNotNull(),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewStringNull().RefineAsNotNull(),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewStringUnknown(),
			expectedFound: false,
		},
		"unknown-with-notnull-refinement": {
			input:           NewStringUnknown().RefineAsNotNull(),
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

func TestStringValue_PrefixRefinement(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input           StringValue
		expectedRefnVal refinement.Refinement
		expectedFound   bool
	}
	tests := map[string]testCase{
		"known-ignored": {
			input:         NewStringValue("test").RefineWithPrefix("hello://"),
			expectedFound: false,
		},
		"null-ignored": {
			input:         NewStringNull().RefineWithPrefix("hello://"),
			expectedFound: false,
		},
		"unknown-no-refinement": {
			input:         NewStringUnknown(),
			expectedFound: false,
		},
		"unknown-with-empty-prefix-refinement": {
			input:         NewStringUnknown().RefineWithPrefix(""),
			expectedFound: false,
		},
		"unknown-with-prefix-refinement": {
			input:           NewStringUnknown().RefineWithPrefix("hello://"),
			expectedRefnVal: refinement.NewStringPrefix("hello://"),
			expectedFound:   true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, found := test.input.PrefixRefinement()
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
