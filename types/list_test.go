package types

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestListTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    ListType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"list-of-strings": {
			input: ListType{
				ElemType: StringType,
			},
			expected: tftypes.List{
				ElementType: tftypes.String,
			},
		},
		"list-of-list-of-strings": {
			input: ListType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
			expected: tftypes.List{
				ElementType: tftypes.List{
					ElementType: tftypes.String,
				},
			},
		},
		"list-of-list-of-list-of-strings": {
			input: ListType{
				ElemType: ListType{
					ElemType: ListType{
						ElemType: StringType,
					},
				},
			},
			expected: tftypes.List{
				ElementType: tftypes.List{
					ElementType: tftypes.List{
						ElementType: tftypes.String,
					},
				},
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := test.input.TerraformType(context.Background())
			if !got.Equal(test.expected) {
				t.Errorf("Expected %s, got %s", test.expected, got)
			}
		})
	}
}

func TestListTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    ListType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"list-of-strings": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			expected: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
		},
		"unknown-list": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: List{
				ElemType: StringType,
				Unknown:  true,
			},
		},
		"partially-unknown-list": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			expected: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
		},
		"null-list": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, nil),
			expected: List{
				ElemType: StringType,
				Null:     true,
			},
		},
		"partially-null-list": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, nil),
			}),
			expected: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
		},
		"wrong-type": {
			receiver: ListType{
				ElemType: StringType,
			},
			input:       tftypes.NewValue(tftypes.String, "wrong"),
			expectedErr: `can't use tftypes.String<"wrong"> as value of List with ElementType types.primitive, can only use tftypes.String values`,
		},
		"wrong-element-type": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.Number,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 1),
			}),
			expectedErr: `can't use tftypes.List[tftypes.Number]<tftypes.Number<"1">> as value of List with ElementType types.primitive, can only use tftypes.String values`,
		},
		"nil-type": {
			receiver: ListType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(nil, nil),
			expected: List{
				ElemType: StringType,
				Null:     true,
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got, gotErr := test.receiver.ValueFromTerraform(context.Background(), test.input)
			if gotErr != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", gotErr.Error())
					return
				}
				if gotErr.Error() != test.expectedErr {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, gotErr.Error())
					return
				}
			}
			if gotErr == nil && test.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", test.expectedErr)
				return
			}
			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("Unexpected diff (-expected, +got): %s", diff)
			}
			if test.expected != nil && test.expected.IsNull() != test.input.IsNull() {
				t.Errorf("Expected null-ness match: expected %t, got %t", test.expected.IsNull(), test.input.IsNull())
			}
			if test.expected != nil && test.expected.IsUnknown() != !test.input.IsKnown() {
				t.Errorf("Expected unknown-ness match: expected %t, got %t", test.expected.IsUnknown(), !test.input.IsKnown())
			}
		})
	}
}

func TestListTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver ListType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: ListType{ElemType: StringType},
			input:    ListType{ElemType: StringType},
			expected: true,
		},
		"diff": {
			receiver: ListType{ElemType: StringType},
			input:    ListType{ElemType: NumberType},
			expected: false,
		},
		"wrongType": {
			receiver: ListType{ElemType: StringType},
			input:    NumberType,
			expected: false,
		},
		"nil": {
			receiver: ListType{ElemType: StringType},
			input:    nil,
			expected: false,
		},
		"nil-elem": {
			receiver: ListType{},
			input:    ListType{},
			// ListTypes with nil ElemTypes are invalid, and
			// aren't equal to anything
			expected: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if test.expected != got {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestListValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      []attr.Value
		expected      List
		expectedDiags diag.Diagnostics
	}{
		"valid-no-elements": {
			elementType: StringType,
			elements:    []attr.Value{},
			expected:    ListValueMust(StringType, []attr.Value{}),
		},
		"valid-elements": {
			elementType: StringType,
			elements: []attr.Value{
				StringNull(),
				StringUnknown(),
				StringValue("test"),
			},
			expected: ListValueMust(
				StringType,
				[]attr.Value{
					StringNull(),
					StringUnknown(),
					StringValue("test"),
				},
			),
		},
		"invalid-element-type": {
			elementType: StringType,
			elements: []attr.Value{
				StringValue("test"),
				BoolValue(true),
			},
			expected: ListUnknown(StringType),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid List Element Type",
					"While creating a List value, an invalid element was detected. "+
						"A List must use the single, given element type. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"List Element Type: types.StringType\n"+
						"List Index (1) Element Type: types.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := ListValue(testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestListValue_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	knownList := ListValueMust(StringType, []attr.Value{StringValue("test")})

	knownList.Null = true

	if knownList.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	knownList.Unknown = true

	if knownList.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	knownList.Elems = []attr.Value{StringValue("not-test")}

	if knownList.Elements()[0].Equal(StringValue("not-test")) {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestListNull_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	nullList := ListNull(StringType)

	nullList.Null = false

	if !nullList.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	nullList.Unknown = true

	if nullList.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	nullList.Elems = []attr.Value{StringValue("test")}

	if len(nullList.Elements()) > 0 {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestListUnknown_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	unknownList := ListUnknown(StringType)

	unknownList.Null = true

	if unknownList.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	unknownList.Unknown = false

	if !unknownList.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	unknownList.Elems = []attr.Value{StringValue("test")}

	if len(unknownList.Elements()) > 0 {
		t.Error("unexpected value update after Value field setting")
	}
}

func TestListElementsAs_stringSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []string
	expected := []string{"hello", "world"}

	diags := (List{
		ElemType: StringType,
		Elems: []attr.Value{
			String{Value: "hello"},
			String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestListElementsAs_attributeValueSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []String
	expected := []String{
		{Value: "hello"},
		{Value: "world"},
	}

	diags := (List{
		ElemType: StringType,
		Elems: []attr.Value{
			String{Value: "hello"},
			String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestListToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       List
		expectation tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"known": {
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"known-partial-unknown": {
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringUnknown(),
					StringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"known-partial-null": {
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringNull(),
					StringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"unknown": {
			input:       ListUnknown(StringType),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"null": {
			input:       ListNull(StringType),
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		},
		"deprecated-known": {
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"deprecated-known-duplicates": {
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "hello"},
				},
			},
			// Duplicate validation does not occur during this method.
			// This is okay, as tftypes allows duplicates.
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "hello"),
			}),
		},
		"deprecated-unknown": {
			input:       List{ElemType: StringType, Unknown: true},
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       List{ElemType: StringType, Null: true},
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		},
		"deprecated-known-partial-unknown": {
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Unknown: true},
					String{Value: "hello, world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"deprecated-known-partial-null": {
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Null: true},
					String{Value: "hello, world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"no-elem-type": {
			input: List{
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: tftypes.Value{},
			expectedErr: "cannot convert List to tftypes.Value if ElemType field is not set",
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

func TestListElements(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    List
		expected []attr.Value
	}{
		"known": {
			input:    ListValueMust(StringType, []attr.Value{StringValue("test")}),
			expected: []attr.Value{StringValue("test")},
		},
		"deprecated-known": {
			input:    List{ElemType: StringType, Elems: []attr.Value{StringValue("test")}},
			expected: []attr.Value{StringValue("test")},
		},
		"null": {
			input:    ListNull(StringType),
			expected: nil,
		},
		"deprecated-null": {
			input:    List{ElemType: StringType, Null: true},
			expected: nil,
		},
		"unknown": {
			input:    ListUnknown(StringType),
			expected: nil,
		},
		"deprecated-unknown": {
			input:    List{ElemType: StringType, Unknown: true},
			expected: nil,
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

func TestListElementType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    List
		expected attr.Type
	}{
		"known": {
			input:    ListValueMust(StringType, []attr.Value{StringValue("test")}),
			expected: StringType,
		},
		"deprecated-known": {
			input:    List{ElemType: StringType, Elems: []attr.Value{StringValue("test")}},
			expected: StringType,
		},
		"null": {
			input:    ListNull(StringType),
			expected: StringType,
		},
		"deprecated-null": {
			input:    List{ElemType: StringType, Null: true},
			expected: StringType,
		},
		"unknown": {
			input:    ListUnknown(StringType),
			expected: StringType,
		},
		"deprecated-unknown": {
			input:    List{ElemType: StringType, Unknown: true},
			expected: StringType,
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

func TestListEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver List
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expected: true,
		},
		"known-known-diff-value": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("goodnight"),
					StringValue("moon"),
				},
			),
			expected: false,
		},
		"known-known-diff-length": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
					StringValue("extra"),
				},
			),
			expected: false,
		},
		"known-known-diff-type": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input: SetValueMust(
				BoolType,
				[]attr.Value{
					BoolValue(false),
					BoolValue(true),
				},
			),
			expected: false,
		},
		"known-known-diff-unknown": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringUnknown(),
				},
			),
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expected: false,
		},
		"known-known-diff-null": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringNull(),
				},
			),
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expected: false,
		},
		"known-unknown": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    ListUnknown(StringType),
			expected: false,
		},
		"known-null": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    ListNull(StringType),
			expected: false,
		},
		"known-diff-type": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expected: false,
		},
		"known-nil": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    nil,
			expected: false,
		},
		"known-deprecated-known": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false, // intentional
		},
		"known-deprecated-unknown": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    List{ElemType: StringType, Unknown: true},
			expected: false,
		},
		"known-deprecated-null": {
			receiver: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    List{ElemType: StringType, Null: true},
			expected: false,
		},
		"deprecated-known-deprecated-known": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: true,
		},
		"deprecated-known-deprecated-known-diff-value": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "goodnight"},
					String{Value: "moon"},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-length": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
					String{Value: "test"},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-type": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: List{
				ElemType: BoolType,
				Elems: []attr.Value{
					Bool{Value: false},
					Bool{Value: true},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-unknown": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-null": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-unknown": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    List{Unknown: true},
			expected: false,
		},
		"deprecated-known-deprecated-null": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    List{Null: true},
			expected: false,
		},
		"deprecated-known-known": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expected: false, // intentional
		},
		"deprecated-known-unknown": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    ListUnknown(StringType),
			expected: false,
		},
		"deprecated-known-null": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    ListNull(StringType),
			expected: false,
		},
		"deprecated-known-diff-type": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"deprecated-known-nil": {
			receiver: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
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

func TestListIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    List
		expected bool
	}{
		"known": {
			input:    ListValueMust(StringType, []attr.Value{StringValue("test")}),
			expected: false,
		},
		"deprecated-known": {
			input:    List{ElemType: StringType, Elems: []attr.Value{StringValue("test")}},
			expected: false,
		},
		"null": {
			input:    ListNull(StringType),
			expected: true,
		},
		"deprecated-null": {
			input:    List{ElemType: StringType, Null: true},
			expected: true,
		},
		"unknown": {
			input:    ListUnknown(StringType),
			expected: false,
		},
		"deprecated-unknown": {
			input:    List{ElemType: StringType, Unknown: true},
			expected: false,
		},
		"deprecated-invalid": {
			input:    List{ElemType: StringType, Null: true, Unknown: true},
			expected: true,
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

func TestListIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    List
		expected bool
	}{
		"known": {
			input:    ListValueMust(StringType, []attr.Value{StringValue("test")}),
			expected: false,
		},
		"deprecated-known": {
			input:    List{ElemType: StringType, Elems: []attr.Value{StringValue("test")}},
			expected: false,
		},
		"null": {
			input:    ListNull(StringType),
			expected: false,
		},
		"deprecated-null": {
			input:    List{ElemType: StringType, Null: true},
			expected: false,
		},
		"unknown": {
			input:    ListUnknown(StringType),
			expected: true,
		},
		"deprecated-unknown": {
			input:    List{ElemType: StringType, Unknown: true},
			expected: true,
		},
		"deprecated-invalid": {
			input:    List{ElemType: StringType, Null: true, Unknown: true},
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

func TestListString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       List
		expectation string
	}
	tests := map[string]testCase{
		"known": {
			input: ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expectation: `["hello","world"]`,
		},
		"known-list-of-lists": {
			input: ListValueMust(
				ListType{
					ElemType: StringType,
				},
				[]attr.Value{
					ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
					ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("foo"),
							StringValue("bar"),
						},
					),
				},
			),
			expectation: `[["hello","world"],["foo","bar"]]`,
		},
		"unknown": {
			input:       ListUnknown(StringType),
			expectation: "<unknown>",
		},
		"null": {
			input:       ListNull(StringType),
			expectation: "<null>",
		},
		"deprecated-known": {
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: `["hello","world"]`,
		},
		"deprecated-known-list-of-lists": {
			input: List{
				ElemType: ListType{
					ElemType: StringType,
				},
				Elems: []attr.Value{
					List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "foo"},
							String{Value: "bar"},
						},
					},
				},
			},
			expectation: `[["hello","world"],["foo","bar"]]`,
		},
		"deprecated-unknown": {
			input:       List{Unknown: true},
			expectation: "<unknown>",
		},
		"deprecated-null": {
			input:       List{Null: true},
			expectation: "<null>",
		},
		"default-empty": {
			input:       List{},
			expectation: "[]",
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

func TestListType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       List
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: ListValueMust(
				StringType,
				[]attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			),
			expectation: ListType{ElemType: StringType},
		},
		"known-list-of-lists": {
			input: ListValueMust(
				ListType{
					ElemType: StringType,
				},
				[]attr.Value{
					ListValueMust(
						StringType,
						[]attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					),
					ListValueMust(
						StringType,
						[]attr.Value{
							String{Value: "foo"},
							String{Value: "bar"},
						},
					),
				},
			),
			expectation: ListType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
		},
		"unknown": {
			input:       ListUnknown(StringType),
			expectation: ListType{ElemType: StringType},
		},
		"null": {
			input:       ListNull(StringType),
			expectation: ListType{ElemType: StringType},
		},
		"deprecated-known": {
			input: List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: ListType{ElemType: StringType},
		},
		"deprecated-known-list-of-lists": {
			input: List{
				ElemType: ListType{
					ElemType: StringType,
				},
				Elems: []attr.Value{
					List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "foo"},
							String{Value: "bar"},
						},
					},
				},
			},
			expectation: ListType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
		},
		"deprecated-unknown": {
			input:       List{ElemType: StringType, Unknown: true},
			expectation: ListType{ElemType: StringType},
		},
		"deprecated-null": {
			input:       List{ElemType: StringType, Null: true},
			expectation: ListType{ElemType: StringType},
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
				ElemType: StringType,
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
				ElemType: StringType,
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
