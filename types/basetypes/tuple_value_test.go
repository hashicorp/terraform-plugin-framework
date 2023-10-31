// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestNewTupleValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementTypes  []attr.Type
		elements      []attr.Value
		expected      TupleValue
		expectedDiags diag.Diagnostics
	}{
		"valid-elements": {
			elementTypes: []attr.Type{StringType{}, BoolType{}, NumberType{}},
			elements: []attr.Value{
				NewStringNull(),
				NewBoolValue(true),
				NewNumberUnknown(),
			},
			expected: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}, NumberType{}},
				[]attr.Value{
					NewStringNull(),
					NewBoolValue(true),
					NewNumberUnknown(),
				},
			),
		},
		"valid-no-elements-or-types": {
			elementTypes: []attr.Type{},
			elements:     []attr.Value{},
			expected:     NewTupleValueMust([]attr.Type{}, []attr.Value{}),
		},
		"invalid-no-elements": {
			elementTypes: []attr.Type{StringType{}, BoolType{}},
			elements:     []attr.Value{},
			expected:     NewTupleUnknown([]attr.Type{StringType{}, BoolType{}}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Tuple Elements",
					"While creating a Tuple value, mismatched element types were detected. "+
						"A Tuple must be an ordered array of elements where the values exactly match the length and types of the defined element types. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Tuple Expected Type: [basetypes.StringType basetypes.BoolType]\n"+
						"Tuple Given Type: []",
				),
			},
		},
		"invalid-mismatched-length": {
			elementTypes: []attr.Type{StringType{}},
			elements:     []attr.Value{NewStringValue("hello"), NewBoolValue(true)},
			expected:     NewTupleUnknown([]attr.Type{StringType{}}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Tuple Elements",
					"While creating a Tuple value, mismatched element types were detected. "+
						"A Tuple must be an ordered array of elements where the values exactly match the length and types of the defined element types. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Tuple Expected Type: [basetypes.StringType]\n"+
						"Tuple Given Type: [basetypes.StringType basetypes.BoolType]",
				),
			},
		},
		"invalid-mismatched-element-types": {
			elementTypes: []attr.Type{BoolType{}, StringType{}},
			elements:     []attr.Value{NewStringValue("hello"), NewBoolValue(true)},
			expected:     NewTupleUnknown([]attr.Type{BoolType{}, StringType{}}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Tuple Element",
					"While creating a Tuple value, an invalid element was detected. "+
						"A Tuple must be an ordered array of elements where the values exactly match the length and types of the defined element types. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Tuple Index (0) Expected Type: basetypes.BoolType\n"+
						"Tuple Index (0) Given Type: basetypes.StringType",
				),
				diag.NewErrorDiagnostic(
					"Invalid Tuple Element",
					"While creating a Tuple value, an invalid element was detected. "+
						"A Tuple must be an ordered array of elements where the values exactly match the length and types of the defined element types. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Tuple Index (1) Expected Type: basetypes.StringType\n"+
						"Tuple Index (1) Given Type: basetypes.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := NewTupleValue(testCase.elementTypes, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestTupleValueElements(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    TupleValue
		expected []attr.Value
	}{
		"known": {
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}},
				[]attr.Value{NewStringValue("test"), NewBoolValue(true)},
			),
			expected: []attr.Value{NewStringValue("test"), NewBoolValue(true)},
		},
		"null": {
			input:    NewTupleNull([]attr.Type{StringType{}, BoolType{}}),
			expected: []attr.Value{},
		},
		"unknown": {
			input:    NewTupleUnknown([]attr.Type{StringType{}, BoolType{}}),
			expected: []attr.Value{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.Elements()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestTupleValueElements_immutable(t *testing.T) {
	t.Parallel()

	value := NewTupleValueMust([]attr.Type{StringType{}}, []attr.Value{NewStringValue("original")})
	value.Elements()[0] = NewStringValue("modified")

	if !value.Equal(NewTupleValueMust([]attr.Type{StringType{}}, []attr.Value{NewStringValue("original")})) {
		t.Fatal("unexpected Elements mutation")
	}
}

func TestTupleValueElementType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    TupleValue
		expected []attr.Type
	}{
		"known": {
			input:    NewTupleValueMust([]attr.Type{StringType{}, BoolType{}}, []attr.Value{NewStringValue("test"), NewBoolValue(true)}),
			expected: []attr.Type{StringType{}, BoolType{}},
		},
		"null": {
			input:    NewTupleNull([]attr.Type{StringType{}, BoolType{}}),
			expected: []attr.Type{StringType{}, BoolType{}},
		},
		"unknown": {
			input:    NewTupleUnknown([]attr.Type{StringType{}, BoolType{}}),
			expected: []attr.Type{StringType{}, BoolType{}},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ElementTypes(context.Background())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

// TODO: Equal

func TestTupleValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    TupleValue
		expected bool
	}{
		"known": {
			input:    NewTupleValueMust([]attr.Type{StringType{}, BoolType{}}, []attr.Value{NewStringValue("test"), NewBoolValue(true)}),
			expected: false,
		},
		"null": {
			input:    NewTupleNull([]attr.Type{StringType{}, BoolType{}}),
			expected: true,
		},
		"unknown": {
			input:    NewTupleUnknown([]attr.Type{StringType{}, BoolType{}}),
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

func TestTupleValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    TupleValue
		expected bool
	}{
		"known": {
			input:    NewTupleValueMust([]attr.Type{StringType{}, BoolType{}}, []attr.Value{NewStringValue("test"), NewBoolValue(true)}),
			expected: false,
		},
		"null": {
			input:    NewTupleNull([]attr.Type{StringType{}, BoolType{}}),
			expected: false,
		},
		"unknown": {
			input:    NewTupleUnknown([]attr.Type{StringType{}, BoolType{}}),
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

// TODO: String tests

func TestTupleValueType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       TupleValue
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, StringType{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expectation: TupleType{ElemTypes: []attr.Type{StringType{}, StringType{}}},
		},
		"known-tuple-of-tuples": {
			input: NewTupleValueMust(
				[]attr.Type{
					TupleType{
						ElemTypes: []attr.Type{StringType{}, StringType{}},
					},
					TupleType{
						ElemTypes: []attr.Type{BoolType{}, BoolType{}},
					},
				},
				[]attr.Value{
					NewTupleValueMust(
						[]attr.Type{StringType{}, StringType{}},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					NewTupleValueMust(
						[]attr.Type{BoolType{}, BoolType{}},
						[]attr.Value{
							NewBoolValue(true),
							NewBoolValue(false),
						},
					),
				},
			),
			expectation: TupleType{
				ElemTypes: []attr.Type{
					TupleType{
						ElemTypes: []attr.Type{StringType{}, StringType{}},
					},
					TupleType{
						ElemTypes: []attr.Type{BoolType{}, BoolType{}},
					},
				},
			},
		},
		"unknown": {
			input:       NewTupleUnknown([]attr.Type{StringType{}, NumberType{}}),
			expectation: TupleType{ElemTypes: []attr.Type{StringType{}, NumberType{}}},
		},
		"null": {
			input:       NewTupleNull([]attr.Type{StringType{}, NumberType{}}),
			expectation: TupleType{ElemTypes: []attr.Type{StringType{}, NumberType{}}},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.Type(context.Background())
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %q, got %q", test.expectation, got)
			}
		})
	}
}

// TODO: ToTerraformValue tests
