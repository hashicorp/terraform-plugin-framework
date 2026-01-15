// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ElementTypes(context.Background())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestTupleValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver TupleValue
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			expected: true,
		},
		"known-known-empty": {
			receiver: NewTupleValueMust(
				[]attr.Type{},
				[]attr.Value{},
			),
			input: NewTupleValueMust(
				[]attr.Type{},
				[]attr.Value{},
			),
			expected: true,
		},
		"known-known-diff-value": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(67890),
				},
			),
			expected: false,
		},
		"known-known-diff-type": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input: NewTupleValueMust(
				[]attr.Type{Int64Type{}, StringType{}},
				[]attr.Value{
					NewInt64Value(12345),
					NewStringValue("hello"),
				},
			),
			expected: false,
		},
		"known-known-diff-type-length": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
					NewInt64Value(67890),
				},
			),
			expected: false,
		},
		// This test just checks there are no panics if an invalid TupleType/Value is defined
		"known-known-diff-element-length": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input: TupleValue{
				state:        attr.ValueStateKnown,
				elementTypes: []attr.Type{StringType{}, Int64Type{}},
				elements: []attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
					NewInt64Value(67890),
				},
			},
			expected: false,
		},
		"known-known-diff-unknown": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Unknown(),
				},
			),
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			expected: false,
		},
		"known-known-diff-null": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Null(),
				},
			),
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			expected: false,
		},
		"known-unknown": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input:    NewTupleUnknown([]attr.Type{StringType{}, Int64Type{}}),
			expected: false,
		},
		"known-null": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input:    NewTupleNull([]attr.Type{StringType{}, Int64Type{}}),
			expected: false,
		},
		"known-diff-type": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expected: false,
		},
		"known-nil": {
			receiver: NewTupleValueMust(
				[]attr.Type{StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewInt64Value(12345),
				},
			),
			input:    nil,
			expected: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if got != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsUnknown()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestTupleValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       TupleValue
		expectation string
	}
	tests := map[string]testCase{
		"known-empty": {
			input: NewTupleValueMust(
				[]attr.Type{},
				[]attr.Value{},
			),
			expectation: `[]`,
		},
		"known": {
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}, StringType{}, Int64Type{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewBoolValue(true),
					NewStringValue("world"),
					NewInt64Value(123),
				},
			),
			expectation: `["hello",true,"world",123]`,
		},
		"known-tuple-of-tuples": {
			input: NewTupleValueMust(
				[]attr.Type{
					TupleType{
						ElemTypes: []attr.Type{StringType{}, BoolType{}},
					},
					TupleType{
						ElemTypes: []attr.Type{Int64Type{}, ObjectType{AttrTypes: map[string]attr.Type{"testattr": StringType{}}}},
					},
				},
				[]attr.Value{
					NewTupleValueMust(
						[]attr.Type{StringType{}, BoolType{}},
						[]attr.Value{
							NewStringValue("hello"),
							NewBoolValue(true),
						},
					),
					NewTupleValueMust(
						[]attr.Type{Int64Type{}, ObjectType{AttrTypes: map[string]attr.Type{"testattr": StringType{}}}},
						[]attr.Value{
							NewInt64Value(1234),
							NewObjectValueMust(
								map[string]attr.Type{"testattr": StringType{}},
								map[string]attr.Value{"testattr": NewStringValue("world")},
							),
						},
					),
				},
			),
			expectation: `[["hello",true],[1234,{"testattr":"world"}]]`,
		},
		"unknown": {
			input:       NewTupleUnknown([]attr.Type{StringType{}, BoolType{}}),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewTupleNull([]attr.Type{StringType{}, BoolType{}}),
			expectation: "<null>",
		},
		"zero-value": {
			input:       TupleValue{},
			expectation: "<null>",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.String()
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %q, got %q", test.expectation, got)
			}
		})
	}
}

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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.Type(context.Background())
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %q, got %q", test.expectation, got)
			}
		})
	}
}

func TestTupleValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       TupleValue
		expectation tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"known": {
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}, DynamicType{}, DynamicType{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewBoolValue(true),
					NewDynamicValue(NewStringValue("world")),
					NewDynamicValue(NewBoolValue(false)),
				},
			),
			expectation: tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType, tftypes.DynamicPseudoType}}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.String, "world"),
				tftypes.NewValue(tftypes.Bool, false),
			}),
		},
		"known-partial-unknown": {
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}, DynamicType{}, DynamicType{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewBoolUnknown(),
					NewDynamicValue(NewStringValue("world")),
					NewDynamicUnknown(),
				},
			),
			expectation: tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType, tftypes.DynamicPseudoType}}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "world"),
				tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
			}),
		},
		"known-partial-null": {
			input: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}, DynamicType{}, DynamicType{}},
				[]attr.Value{
					NewStringNull(),
					NewBoolValue(true),
					NewDynamicValue(NewStringValue("world")),
					NewDynamicNull(),
				},
			),
			expectation: tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType, tftypes.DynamicPseudoType}}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.String, "world"),
				tftypes.NewValue(tftypes.DynamicPseudoType, nil),
			}),
		},
		"unknown": {
			input:       NewTupleUnknown([]attr.Type{StringType{}, BoolType{}, DynamicType{}}),
			expectation: tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType}}, tftypes.UnknownValue),
		},
		"null": {
			input:       NewTupleNull([]attr.Type{StringType{}, BoolType{}, DynamicType{}}),
			expectation: tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType}}, nil),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := test.input.ToTerraformValue(context.Background())

			if test.expectedErr == "" && gotErr != nil {
				t.Errorf("Unexpected error: %s", gotErr)
				return
			}

			if test.expectedErr != "" {
				if gotErr == nil {
					t.Errorf("Expected error to be %q, got none", test.expectedErr)
					return
				}

				if test.expectedErr != gotErr.Error() {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, gotErr.Error())
					return
				}
			}

			if diff := cmp.Diff(got, test.expectation); diff != "" {
				t.Errorf("Unexpected result (+got, -expected): %s", diff)
			}
		})
	}
}
