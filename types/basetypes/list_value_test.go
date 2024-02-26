// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestNewListValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      []attr.Value
		expected      ListValue
		expectedDiags diag.Diagnostics
	}{
		"valid-no-elements": {
			elementType: StringType{},
			elements:    []attr.Value{},
			expected:    NewListValueMust(StringType{}, []attr.Value{}),
		},
		"valid-elements": {
			elementType: StringType{},
			elements: []attr.Value{
				NewStringNull(),
				NewStringUnknown(),
				NewStringValue("test"),
			},
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringNull(),
					NewStringUnknown(),
					NewStringValue("test"),
				},
			),
		},
		"invalid-element-type": {
			elementType: StringType{},
			elements: []attr.Value{
				NewStringValue("test"),
				NewBoolValue(true),
			},
			expected: NewListUnknown(StringType{}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid List Element Type",
					"While creating a List value, an invalid element was detected. "+
						"A List must use the single, given element type. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"List Element Type: basetypes.StringType\n"+
						"List Index (1) Element Type: basetypes.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := NewListValue(testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestNewListValueFrom(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      any
		expected      ListValue
		expectedDiags diag.Diagnostics
	}{
		"valid-StringType{}-[]attr.Value-empty": {
			elementType: StringType{},
			elements:    []attr.Value{},
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{},
			),
		},
		"valid-StringType{}-[]types.String-empty": {
			elementType: StringType{},
			elements:    []StringValue{},
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{},
			),
		},
		"valid-StringType{}-[]types.String": {
			elementType: StringType{},
			elements: []StringValue{
				NewStringNull(),
				NewStringUnknown(),
				NewStringValue("test"),
			},
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringNull(),
					NewStringUnknown(),
					NewStringValue("test"),
				},
			),
		},
		"valid-StringType{}-[]*string": {
			elementType: StringType{},
			elements: []*string{
				nil,
				pointer("test1"),
				pointer("test2"),
			},
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringNull(),
					NewStringValue("test1"),
					NewStringValue("test2"),
				},
			),
		},
		"valid-StringType{}-[]string": {
			elementType: StringType{},
			elements: []string{
				"test1",
				"test2",
			},
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("test1"),
					NewStringValue("test2"),
				},
			),
		},
		"invalid-not-slice": {
			elementType: StringType{},
			elements:    "oops",
			expected:    NewListUnknown(StringType{}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"List Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected List value, received tftypes.Value with value: tftypes.String<\"oops\">",
				),
			},
		},
		"invalid-type": {
			elementType: StringType{},
			elements:    []bool{true},
			expected:    NewListUnknown(StringType{}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtListIndex(0),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"can't unmarshal tftypes.Bool into *string, expected string",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := NewListValueFrom(context.Background(), testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestListElementsAs_stringSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []string
	expected := []string{"hello", "world"}

	diags := NewListValueMust(
		StringType{},
		[]attr.Value{
			NewStringValue("hello"),
			NewStringValue("world"),
		},
	).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestListElementsAs_attributeValueSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []StringValue
	expected := []StringValue{
		NewStringValue("hello"),
		NewStringValue("world"),
	}

	diags := NewListValueMust(
		StringType{},
		[]attr.Value{
			NewStringValue("hello"),
			NewStringValue("world"),
		},
	).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestListValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       ListValue
		expectation tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"known": {
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"known-partial-unknown": {
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringUnknown(),
					NewStringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"known-partial-null": {
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringNull(),
					NewStringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"unknown": {
			input:       NewListUnknown(StringType{}),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"null": {
			input:       NewListNull(StringType{}),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		},
	}
	for name, test := range tests {
		name, test := name, test
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

func TestListValueElements(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ListValue
		expected []attr.Value
	}{
		"known": {
			input:    NewListValueMust(StringType{}, []attr.Value{NewStringValue("test")}),
			expected: []attr.Value{NewStringValue("test")},
		},
		"null": {
			input:    NewListNull(StringType{}),
			expected: []attr.Value{},
		},
		"unknown": {
			input:    NewListUnknown(StringType{}),
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

func TestListValueElements_immutable(t *testing.T) {
	t.Parallel()

	value := NewListValueMust(StringType{}, []attr.Value{NewStringValue("original")})
	value.Elements()[0] = NewStringValue("modified")

	if !value.Equal(NewListValueMust(StringType{}, []attr.Value{NewStringValue("original")})) {
		t.Fatal("unexpected Elements mutation")
	}
}

func TestListValueElementType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ListValue
		expected attr.Type
	}{
		"known": {
			input:    NewListValueMust(StringType{}, []attr.Value{NewStringValue("test")}),
			expected: StringType{},
		},
		"null": {
			input:    NewListNull(StringType{}),
			expected: StringType{},
		},
		"unknown": {
			input:    NewListUnknown(StringType{}),
			expected: StringType{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ElementType(context.Background())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestListValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver ListValue
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expected: true,
		},
		"known-known-diff-value": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("goodnight"),
					NewStringValue("moon"),
				},
			),
			expected: false,
		},
		"known-known-diff-length": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
					NewStringValue("extra"),
				},
			),
			expected: false,
		},
		"known-known-diff-type": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input: NewSetValueMust(
				BoolType{},
				[]attr.Value{
					NewBoolValue(false),
					NewBoolValue(true),
				},
			),
			expected: false,
		},
		"known-known-diff-unknown": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringUnknown(),
				},
			),
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expected: false,
		},
		"known-known-diff-null": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringNull(),
				},
			),
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expected: false,
		},
		"known-unknown": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input:    NewListUnknown(StringType{}),
			expected: false,
		},
		"known-null": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input:    NewListNull(StringType{}),
			expected: false,
		},
		"known-diff-type": {
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
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
			receiver: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input:    nil,
			expected: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if got != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestListValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ListValue
		expected bool
	}{
		"known": {
			input:    NewListValueMust(StringType{}, []attr.Value{NewStringValue("test")}),
			expected: false,
		},
		"null": {
			input:    NewListNull(StringType{}),
			expected: true,
		},
		"unknown": {
			input:    NewListUnknown(StringType{}),
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

func TestListValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ListValue
		expected bool
	}{
		"known": {
			input:    NewListValueMust(StringType{}, []attr.Value{NewStringValue("test")}),
			expected: false,
		},
		"null": {
			input:    NewListNull(StringType{}),
			expected: false,
		},
		"unknown": {
			input:    NewListUnknown(StringType{}),
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

func TestListValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       ListValue
		expectation string
	}
	tests := map[string]testCase{
		"known": {
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expectation: `["hello","world"]`,
		},
		"known-list-of-lists": {
			input: NewListValueMust(
				ListType{
					ElemType: StringType{},
				},
				[]attr.Value{
					NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("foo"),
							NewStringValue("bar"),
						},
					),
				},
			),
			expectation: `[["hello","world"],["foo","bar"]]`,
		},
		"unknown": {
			input:       NewListUnknown(StringType{}),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewListNull(StringType{}),
			expectation: "<null>",
		},
		"zero-value": {
			input:       ListValue{},
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

func TestListValueType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       ListValue
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expectation: ListType{ElemType: StringType{}},
		},
		"known-list-of-lists": {
			input: NewListValueMust(
				ListType{
					ElemType: StringType{},
				},
				[]attr.Value{
					NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("foo"),
							NewStringValue("bar"),
						},
					),
				},
			),
			expectation: ListType{
				ElemType: ListType{
					ElemType: StringType{},
				},
			},
		},
		"unknown": {
			input:       NewListUnknown(StringType{}),
			expectation: ListType{ElemType: StringType{}},
		},
		"null": {
			input:       NewListNull(StringType{}),
			expectation: ListType{ElemType: StringType{}},
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

func TestListTypeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		listType      ListType
		tfValue       tftypes.Value
		path          path.Path
		expectedDiags diag.Diagnostics
	}{
		"wrong-value-type": {
			listType: ListType{
				ElemType: StringType{},
			},
			tfValue: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "testvalue"),
			}),
			path: path.Root("test"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"List Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected List value, received tftypes.Value with value: tftypes.Set[tftypes.String]<tftypes.String<\"testvalue\">>",
				),
			},
		},
		"no-validation": {
			listType: ListType{
				ElemType: StringType{},
			},
			tfValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "testvalue"),
			}),
			path: path.Root("test"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.listType.Validate(context.Background(), testCase.tfValue, testCase.path)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
