// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSetElementsAs_stringSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []string
	expected := []string{"hello", "world"}

	diags := NewSetValueMust(
		StringType{},
		[]attr.Value{
			NewStringValue("hello"),
			NewStringValue("world"),
		},
	).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %s", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestSetElementsAs_attributeValueSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []StringValue
	expected := []StringValue{
		NewStringValue("hello"),
		NewStringValue("world"),
	}

	diags := NewSetValueMust(
		StringType{},
		[]attr.Value{
			NewStringValue("hello"),
			NewStringValue("world"),
		},
	).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %s", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

var benchDiags diag.Diagnostics // Prevent compiler optimization

func benchmarkSetTypeValidate(b *testing.B, elementCount int) {
	elements := make([]tftypes.Value, 0, elementCount)

	for idx := range elements {
		elements[idx] = tftypes.NewValue(tftypes.String, strconv.Itoa(idx))
	}

	var diags diag.Diagnostics // Prevent compiler optimization
	ctx := context.Background()
	in := tftypes.NewValue(
		tftypes.Set{
			ElementType: tftypes.String,
		},
		elements,
	)
	path := path.Root("test")
	set := SetType{}

	for n := 0; n < b.N; n++ {
		diags = set.Validate(ctx, in, path)
	}

	benchDiags = diags
}

func BenchmarkSetTypeValidate10(b *testing.B) {
	benchmarkSetTypeValidate(b, 10)
}

func BenchmarkSetTypeValidate100(b *testing.B) {
	benchmarkSetTypeValidate(b, 100)
}

func BenchmarkSetTypeValidate1000(b *testing.B) {
	benchmarkSetTypeValidate(b, 1000)
}

func BenchmarkSetTypeValidate10000(b *testing.B) {
	benchmarkSetTypeValidate(b, 10000)
}

func BenchmarkSetTypeValidate100000(b *testing.B) {
	benchmarkSetTypeValidate(b, 100000)
}

func BenchmarkSetTypeValidate1000000(b *testing.B) {
	benchmarkSetTypeValidate(b, 1000000)
}

func TestSetTypeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in            tftypes.Value
		expectedDiags diag.Diagnostics
	}{
		"empty-struct": {
			in: tftypes.Value{},
		},
		"null": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				nil,
			),
		},
		"null-element": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, nil),
				},
			),
		},
		"null-elements": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, nil),
					tftypes.NewValue(tftypes.String, nil),
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Duplicate Set Element",
					"This attribute contains duplicate values of: tftypes.String<null>",
				),
			},
		},
		"unknown": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				tftypes.UnknownValue,
			),
		},
		"unknown-element": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
		},
		"unknown-elements": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
		},
		"value": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
				},
			),
		},
		"value-and-null": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, nil),
				},
			),
		},
		"value-and-unknown": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
		},
		"values": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				},
			),
		},
		"values-duplicates": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "hello"),
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Duplicate Set Element",
					"This attribute contains duplicate values of: tftypes.String<\"hello\">",
				),
			},
		},
		"values-duplicates-and-unknowns": {
			in: tftypes.NewValue(
				tftypes.Set{
					ElementType: tftypes.String,
				},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					tftypes.NewValue(tftypes.String, "hello"),
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Duplicate Set Element",
					"This attribute contains duplicate values of: tftypes.String<\"hello\">",
				),
			},
		},
		"wrong-value-type": {
			in: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "testvalue"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Set Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected Set value, received tftypes.Value with value: tftypes.List[tftypes.String]<tftypes.String<\"testvalue\">>",
				),
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := SetType{}.Validate(context.Background(), testCase.in, path.Root("test"))

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (+got, -expected): %s", diff)
			}
		})
	}
}

func TestNewSetValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      []attr.Value
		expected      SetValue
		expectedDiags diag.Diagnostics
	}{
		"valid-no-elements": {
			elementType: StringType{},
			elements:    []attr.Value{},
			expected:    NewSetValueMust(StringType{}, []attr.Value{}),
		},
		"valid-elements": {
			elementType: StringType{},
			elements: []attr.Value{
				NewStringNull(),
				NewStringUnknown(),
				NewStringValue("test"),
			},
			expected: NewSetValueMust(
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
			expected: NewSetUnknown(StringType{}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Set Element Type",
					"While creating a Set value, an invalid element was detected. "+
						"A Set must use the single, given element type. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Set Element Type: basetypes.StringType\n"+
						"Set Index (1) Element Type: basetypes.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := NewSetValue(testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestNewSetValueFrom(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      any
		expected      SetValue
		expectedDiags diag.Diagnostics
	}{
		"valid-StringType{}-[]attr.Value-empty": {
			elementType: StringType{},
			elements:    []attr.Value{},
			expected: NewSetValueMust(
				StringType{},
				[]attr.Value{},
			),
		},
		"valid-StringType{}-[]types.String-empty": {
			elementType: StringType{},
			elements:    []StringValue{},
			expected: NewSetValueMust(
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
			expected: NewSetValueMust(
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
			expected: NewSetValueMust(
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
			expected: NewSetValueMust(
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
			expected:    NewSetUnknown(StringType{}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Set Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected Set value, received tftypes.Value with value: tftypes.String<\"oops\">",
				),
			},
		},
		"invalid-type": {
			elementType: StringType{},
			elements:    []bool{true},
			expected:    NewSetUnknown(StringType{}),
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

			got, diags := NewSetValueFrom(context.Background(), testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestSetValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       SetValue
		expectation tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"known": {
			input: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"known-duplicates": {
			input: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("hello"),
				},
			),
			// Duplicate validation does not occur during this method.
			// This is okay, as tftypes allows duplicates.
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "hello"),
			}),
		},
		"known-partial-unknown": {
			input: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringUnknown(),
					NewStringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"known-partial-null": {
			input: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringNull(),
					NewStringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"unknown": {
			input:       NewSetUnknown(StringType{}),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"null": {
			input:       NewSetNull(StringType{}),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
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

func TestSetValueElements(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    SetValue
		expected []attr.Value
	}{
		"known": {
			input:    NewSetValueMust(StringType{}, []attr.Value{NewStringValue("test")}),
			expected: []attr.Value{NewStringValue("test")},
		},
		"null": {
			input:    NewSetNull(StringType{}),
			expected: []attr.Value{},
		},
		"unknown": {
			input:    NewSetUnknown(StringType{}),
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

func TestSetValueElements_immutable(t *testing.T) {
	t.Parallel()

	value := NewSetValueMust(StringType{}, []attr.Value{NewStringValue("original")})
	value.Elements()[0] = NewStringValue("modified")

	if !value.Equal(NewSetValueMust(StringType{}, []attr.Value{NewStringValue("original")})) {
		t.Fatal("unexpected Elements mutation")
	}
}

func TestSetValueElementType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    SetValue
		expected attr.Type
	}{
		"known": {
			input:    NewSetValueMust(StringType{}, []attr.Value{NewStringValue("test")}),
			expected: StringType{},
		},
		"null": {
			input:    NewSetNull(StringType{}),
			expected: StringType{},
		},
		"unknown": {
			input:    NewSetUnknown(StringType{}),
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

func TestSetValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver SetValue
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: NewSetValueMust(
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
			expected: true,
		},
		"known-known-diff-value": {
			receiver: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("goodnight"),
					NewStringValue("moon"),
				},
			),
			expected: false,
		},
		"known-known-diff-length": {
			receiver: NewSetValueMust(
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
					NewStringValue("extra"),
				},
			),
			expected: false,
		},
		"known-known-diff-type": {
			receiver: NewSetValueMust(
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
			receiver: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringUnknown(),
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
		"known-known-diff-null": {
			receiver: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringNull(),
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
		"known-unknown": {
			receiver: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input:    NewSetUnknown(StringType{}),
			expected: false,
		},
		"known-null": {
			receiver: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			input:    NewSetNull(StringType{}),
			expected: false,
		},
		"known-diff-type": {
			receiver: NewSetValueMust(
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
			expected: false,
		},
		"known-nil": {
			receiver: NewSetValueMust(
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

func TestSetValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    SetValue
		expected bool
	}{
		"known": {
			input:    NewSetValueMust(StringType{}, []attr.Value{NewStringValue("test")}),
			expected: false,
		},
		"null": {
			input:    NewSetNull(StringType{}),
			expected: true,
		},
		"unknown": {
			input:    NewSetUnknown(StringType{}),
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

func TestSetValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    SetValue
		expected bool
	}{
		"known": {
			input:    NewSetValueMust(StringType{}, []attr.Value{NewStringValue("test")}),
			expected: false,
		},
		"null": {
			input:    NewSetNull(StringType{}),
			expected: false,
		},
		"unknown": {
			input:    NewSetUnknown(StringType{}),
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

func TestSetValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       SetValue
		expectation string
	}
	tests := map[string]testCase{
		"known": {
			input: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expectation: `["hello","world"]`,
		},
		"known-set-of-sets": {
			input: NewSetValueMust(
				SetType{
					ElemType: StringType{},
				},
				[]attr.Value{
					NewSetValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					NewSetValueMust(
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
			input:       NewSetUnknown(StringType{}),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewSetNull(StringType{}),
			expectation: "<null>",
		},
		"zero-value": {
			input:       SetValue{},
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

func TestSetValueType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       SetValue
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
			expectation: SetType{ElemType: StringType{}},
		},
		"known-set-of-sets": {
			input: NewSetValueMust(
				SetType{
					ElemType: StringType{},
				},
				[]attr.Value{
					NewSetValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					NewSetValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("foo"),
							NewStringValue("bar"),
						},
					),
				},
			),
			expectation: SetType{
				ElemType: SetType{
					ElemType: StringType{},
				},
			},
		},
		"unknown": {
			input:       NewSetUnknown(StringType{}),
			expectation: SetType{ElemType: StringType{}},
		},
		"null": {
			input:       NewSetNull(StringType{}),
			expectation: SetType{ElemType: StringType{}},
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
