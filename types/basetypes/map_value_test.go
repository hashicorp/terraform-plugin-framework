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

func TestNewMapValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      map[string]attr.Value
		expected      MapValue
		expectedDiags diag.Diagnostics
	}{
		"valid-no-elements": {
			elementType: StringType{},
			elements:    map[string]attr.Value{},
			expected:    NewMapValueMust(StringType{}, map[string]attr.Value{}),
		},
		"valid-elements": {
			elementType: StringType{},
			elements: map[string]attr.Value{
				"null":    NewStringNull(),
				"unknown": NewStringUnknown(),
				"known":   NewStringValue("test"),
			},
			expected: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"null":    NewStringNull(),
					"unknown": NewStringUnknown(),
					"known":   NewStringValue("test"),
				},
			),
		},
		"invalid-element-type": {
			elementType: StringType{},
			elements: map[string]attr.Value{
				"string": NewStringValue("test"),
				"bool":   NewBoolValue(true),
			},
			expected: NewMapUnknown(StringType{}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Map Element Type",
					"While creating a Map value, an invalid element was detected. "+
						"A Map must use the single, given element type. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Map Element Type: basetypes.StringType\n"+
						"Map Key (bool) Element Type: basetypes.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := NewMapValue(testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestNewMapValueFrom(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      any
		expected      MapValue
		expectedDiags diag.Diagnostics
	}{
		"valid-StringType{}-map[string]attr.Value-empty": {
			elementType: StringType{},
			elements:    map[string]attr.Value{},
			expected: NewMapValueMust(
				StringType{},
				map[string]attr.Value{},
			),
		},
		"valid-StringType{}-map[string]types.String-empty": {
			elementType: StringType{},
			elements:    map[string]StringValue{},
			expected: NewMapValueMust(
				StringType{},
				map[string]attr.Value{},
			),
		},
		"valid-StringType{}-map[string]types.String": {
			elementType: StringType{},
			elements: map[string]StringValue{
				"key1": NewStringNull(),
				"key2": NewStringUnknown(),
				"key3": NewStringValue("test"),
			},
			expected: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringNull(),
					"key2": NewStringUnknown(),
					"key3": NewStringValue("test"),
				},
			),
		},
		"valid-StringType{}-map[string]*string": {
			elementType: StringType{},
			elements: map[string]*string{
				"key1": nil,
				"key2": pointer("test1"),
				"key3": pointer("test2"),
			},
			expected: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringNull(),
					"key2": NewStringValue("test1"),
					"key3": NewStringValue("test2"),
				},
			),
		},
		"valid-StringType{}-map[string]string": {
			elementType: StringType{},
			elements: map[string]string{
				"key1": "test1",
				"key2": "test2",
			},
			expected: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("test1"),
					"key2": NewStringValue("test2"),
				},
			),
		},
		"invalid-not-map": {
			elementType: StringType{},
			elements:    "oops",
			expected:    NewMapUnknown(StringType{}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Map Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected Map value, received tftypes.Value with value: tftypes.String<\"oops\">",
				),
			},
		},
		"invalid-type": {
			elementType: StringType{},
			elements:    map[string]bool{"key1": true},
			expected:    NewMapUnknown(StringType{}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtMapKey("key1"),
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

			got, diags := NewMapValueFrom(context.Background(), testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestMapElementsAs_mapStringString(t *testing.T) {
	t.Parallel()

	var stringSlice map[string]string
	expected := map[string]string{
		"h": "hello",
		"w": "world",
	}

	diags := NewMapValueMust(
		StringType{},
		map[string]attr.Value{
			"h": NewStringValue("hello"),
			"w": NewStringValue("world"),
		},
	).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestMapElementsAs_mapStringAttributeValue(t *testing.T) {
	t.Parallel()

	var stringSlice map[string]StringValue
	expected := map[string]StringValue{
		"h": NewStringValue("hello"),
		"w": NewStringValue("world"),
	}

	diags := NewMapValueMust(
		StringType{},
		map[string]attr.Value{
			"h": NewStringValue("hello"),
			"w": NewStringValue("world"),
		},
	).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestMapValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       MapValue
		expectation tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"known": {
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, "hello"),
				"key2": tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"known-partial-unknown": {
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringUnknown(),
					"key2": NewStringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"key2": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"known-partial-null": {
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringNull(),
					"key2": NewStringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, nil),
				"key2": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"unknown": {
			input:       NewMapUnknown(StringType{}),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"null": {
			input:       NewMapNull(StringType{}),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
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

func TestMapValueElements(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    MapValue
		expected map[string]attr.Value
	}{
		"known": {
			input:    NewMapValueMust(StringType{}, map[string]attr.Value{"test-key": NewStringValue("test-value")}),
			expected: map[string]attr.Value{"test-key": NewStringValue("test-value")},
		},
		"null": {
			input:    NewMapNull(StringType{}),
			expected: map[string]attr.Value{},
		},
		"unknown": {
			input:    NewMapUnknown(StringType{}),
			expected: map[string]attr.Value{},
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

func TestMapValueElements_immutable(t *testing.T) {
	t.Parallel()

	value := NewMapValueMust(StringType{}, map[string]attr.Value{"test": NewStringValue("original")})
	value.Elements()["test"] = NewStringValue("modified")

	if !value.Equal(NewMapValueMust(StringType{}, map[string]attr.Value{"test": NewStringValue("original")})) {
		t.Fatal("unexpected Elements mutation")
	}
}

func TestMapValueElementType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    MapValue
		expected attr.Type
	}{
		"known": {
			input:    NewMapValueMust(StringType{}, map[string]attr.Value{"test-key": NewStringValue("test-value")}),
			expected: StringType{},
		},
		"null": {
			input:    NewMapNull(StringType{}),
			expected: StringType{},
		},
		"unknown": {
			input:    NewMapUnknown(StringType{}),
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

func TestMapValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver MapValue
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			expected: true,
		},
		"known-known-diff-value": {
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("goodnight"),
					"key2": NewStringValue("moon"),
				},
			),
			expected: false,
		},
		"known-known-diff-length": {
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
					"key3": NewStringValue("extra"),
				},
			),
			expected: false,
		},
		"known-known-diff-type": {
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
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
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringUnknown(),
				},
			),
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			expected: false,
		},
		"known-known-diff-null": {
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringNull(),
				},
			),
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			expected: false,
		},
		"known-unknown": {
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			input:    NewMapUnknown(StringType{}),
			expected: false,
		},
		"known-null": {
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			input:    NewMapNull(StringType{}),
			expected: false,
		},
		"known-diff-type": {
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
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
			receiver: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
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

func TestMapValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    MapValue
		expected bool
	}{
		"known": {
			input:    NewMapValueMust(StringType{}, map[string]attr.Value{"test-key": NewStringValue("test-value")}),
			expected: false,
		},
		"null": {
			input:    NewMapNull(StringType{}),
			expected: true,
		},
		"unknown": {
			input:    NewMapUnknown(StringType{}),
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

func TestMapValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    MapValue
		expected bool
	}{
		"known": {
			input:    NewMapValueMust(StringType{}, map[string]attr.Value{"test-key": NewStringValue("test-value")}),
			expected: false,
		},
		"null": {
			input:    NewMapNull(StringType{}),
			expected: false,
		},
		"unknown": {
			input:    NewMapUnknown(StringType{}),
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

func TestMapValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       MapValue
		expectation string
	}
	tests := map[string]testCase{
		"known": {
			input: NewMapValueMust(
				Int64Type{},
				map[string]attr.Value{
					"alpha": NewInt64Value(1234),
					"beta":  NewInt64Value(56789),
					"gamma": NewInt64Value(9817),
					"sigma": NewInt64Value(62534),
				},
			),
			expectation: `{"alpha":1234,"beta":56789,"gamma":9817,"sigma":62534}`,
		},
		"known-map-of-maps": {
			input: NewMapValueMust(
				MapType{
					ElemType: StringType{},
				},
				map[string]attr.Value{
					"first": NewMapValueMust(
						StringType{},
						map[string]attr.Value{
							"alpha": NewStringValue("hello"),
							"beta":  NewStringValue("world"),
							"gamma": NewStringValue("foo"),
							"sigma": NewStringValue("bar"),
						},
					),
					"second": NewMapValueMust(
						StringType{},
						map[string]attr.Value{
							"echo": NewStringValue("echo"),
						},
					),
				},
			),
			expectation: `{"first":{"alpha":"hello","beta":"world","gamma":"foo","sigma":"bar"},"second":{"echo":"echo"}}`,
		},
		"known-key-quotes": {
			input: NewMapValueMust(
				BoolType{},
				map[string]attr.Value{
					`testing is "fun"`: NewBoolValue(true),
				},
			),
			expectation: `{"testing is \"fun\"":true}`,
		},
		"unknown": {
			input:       NewMapUnknown(StringType{}),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewMapNull(StringType{}),
			expectation: "<null>",
		},
		"zero-value": {
			input:       MapValue{},
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

func TestMapValueType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       MapValue
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: NewMapValueMust(
				StringType{},
				map[string]attr.Value{
					"key1": NewStringValue("hello"),
					"key2": NewStringValue("world"),
				},
			),
			expectation: MapType{ElemType: StringType{}},
		},
		"known-map-of-maps": {
			input: NewMapValueMust(
				MapType{
					ElemType: StringType{},
				},
				map[string]attr.Value{
					"key1": NewMapValueMust(
						StringType{},
						map[string]attr.Value{
							"key1": NewStringValue("hello"),
							"key2": NewStringValue("world"),
						},
					),
					"key2": NewMapValueMust(
						StringType{},
						map[string]attr.Value{
							"key1": NewStringValue("foo"),
							"key2": NewStringValue("bar"),
						},
					),
				},
			),
			expectation: MapType{
				ElemType: MapType{
					ElemType: StringType{},
				},
			},
		},
		"unknown": {
			input:       NewMapUnknown(StringType{}),
			expectation: MapType{ElemType: StringType{}},
		},
		"null": {
			input:       NewMapNull(StringType{}),
			expectation: MapType{ElemType: StringType{}},
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

func TestMapTypeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		mapType       MapType
		tfValue       tftypes.Value
		path          path.Path
		expectedDiags diag.Diagnostics
	}{
		"wrong-value-type": {
			mapType: MapType{
				ElemType: StringType{},
			},
			tfValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "testvalue"),
			}),
			path: path.Root("test"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Map Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected Map value, received tftypes.Value with value: tftypes.List[tftypes.String]<tftypes.String<\"testvalue\">>",
				),
			},
		},
		"no-validation": {
			mapType: MapType{
				ElemType: StringType{},
			},
			tfValue: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
			}),
			path: path.Root("test"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.mapType.Validate(context.Background(), testCase.tfValue, testCase.path)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
