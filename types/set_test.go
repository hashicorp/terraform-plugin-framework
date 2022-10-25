package types

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

func TestSetTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    SetType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"set-of-strings": {
			input: SetType{
				ElemType: StringType,
			},
			expected: tftypes.Set{
				ElementType: tftypes.String,
			},
		},
		"set-of-set-of-strings": {
			input: SetType{
				ElemType: SetType{
					ElemType: StringType,
				},
			},
			expected: tftypes.Set{
				ElementType: tftypes.Set{
					ElementType: tftypes.String,
				},
			},
		},
		"set-of-set-of-set-of-strings": {
			input: SetType{
				ElemType: SetType{
					ElemType: SetType{
						ElemType: StringType,
					},
				},
			},
			expected: tftypes.Set{
				ElementType: tftypes.Set{
					ElementType: tftypes.Set{
						ElementType: tftypes.String,
					},
				},
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.TerraformType(context.Background())
			if !got.Equal(test.expected) {
				t.Errorf("Expected %s, got %s", test.expected, got)
			}
		})
	}
}

func TestSetTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    SetType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"set-of-strings": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
		},
		"set-of-duplicate-strings": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "hello"),
			}),
			// Duplicate validation does not occur during this method.
			// This is okay, as tftypes allows duplicates.
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "hello"},
				},
			},
		},
		"unknown-set": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: Set{
				ElemType: StringType,
				Unknown:  true,
			},
		},
		"partially-unknown-set": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
		},
		"null-set": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, nil),
			expected: Set{
				ElemType: StringType,
				Null:     true,
			},
		},
		"partially-null-set": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, nil),
			}),
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
		},
		"wrong-type": {
			receiver: SetType{
				ElemType: StringType,
			},
			input:       tftypes.NewValue(tftypes.String, "wrong"),
			expectedErr: `can't use tftypes.String<"wrong"> as value of Set with ElementType types.primitive, can only use tftypes.String values`,
		},
		"wrong-element-type": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.Number,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 1),
			}),
			expectedErr: `can't use tftypes.Set[tftypes.Number]<tftypes.Number<"1">> as value of Set with ElementType types.primitive, can only use tftypes.String values`,
		},
		"nil-type": {
			receiver: SetType{
				ElemType: StringType,
			},
			input: tftypes.NewValue(nil, nil),
			expected: Set{
				ElemType: StringType,
				Null:     true,
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

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

func TestSetTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver SetType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: SetType{ElemType: StringType},
			input:    SetType{ElemType: StringType},
			expected: true,
		},
		"diff": {
			receiver: SetType{ElemType: StringType},
			input:    SetType{ElemType: NumberType},
			expected: false,
		},
		"wrongType": {
			receiver: SetType{ElemType: StringType},
			input:    NumberType,
			expected: false,
		},
		"nil": {
			receiver: SetType{ElemType: StringType},
			input:    nil,
			expected: false,
		},
		"nil-elem": {
			receiver: SetType{},
			input:    SetType{},
			// SetTypes with nil ElemTypes are invalid, and
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

func TestSetElementsAs_stringSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []string
	expected := []string{"hello", "world"}

	diags := (Set{
		ElemType: StringType,
		Elems: []attr.Value{
			String{Value: "hello"},
			String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %s", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestSetElementsAs_attributeValueSlice(t *testing.T) {
	t.Parallel()

	var stringSlice []String
	expected := []String{
		{Value: "hello"},
		{Value: "world"},
	}

	diags := (Set{
		ElemType: StringType,
		Elems: []attr.Value{
			String{Value: "hello"},
			String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
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

func TestSetValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      []attr.Value
		expected      Set
		expectedDiags diag.Diagnostics
	}{
		"valid-no-elements": {
			elementType: StringType,
			elements:    []attr.Value{},
			expected:    SetValueMust(StringType, []attr.Value{}),
		},
		"valid-elements": {
			elementType: StringType,
			elements: []attr.Value{
				StringNull(),
				StringUnknown(),
				StringValue("test"),
			},
			expected: SetValueMust(
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
			expected: SetUnknown(StringType),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Set Element Type",
					"While creating a Set value, an invalid element was detected. "+
						"A Set must use the single, given element type. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Set Element Type: types.StringType\n"+
						"Set Index (1) Element Type: types.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := SetValue(testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestSetValueFrom(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      any
		expected      Set
		expectedDiags diag.Diagnostics
	}{
		"valid-StringType-[]attr.Value-empty": {
			elementType: StringType,
			elements:    []attr.Value{},
			expected: Set{
				ElemType: StringType,
				Elems:    []attr.Value{},
			},
		},
		"valid-StringType-[]types.String-empty": {
			elementType: StringType,
			elements:    []String{},
			expected: Set{
				ElemType: StringType,
				Elems:    []attr.Value{},
			},
		},
		"valid-StringType-[]types.String": {
			elementType: StringType,
			elements: []String{
				StringNull(),
				StringUnknown(),
				StringValue("test"),
			},
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Null: true},
					String{Unknown: true},
					String{Value: "test"},
				},
			},
		},
		"valid-StringType-[]*string": {
			elementType: StringType,
			elements: []*string{
				nil,
				pointer("test1"),
				pointer("test2"),
			},
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Null: true},
					String{Value: "test1"},
					String{Value: "test2"},
				},
			},
		},
		"valid-StringType-[]string": {
			elementType: StringType,
			elements: []string{
				"test1",
				"test2",
			},
			expected: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "test1"},
					String{Value: "test2"},
				},
			},
		},
		"invalid-not-slice": {
			elementType: StringType,
			elements:    "oops",
			expected:    SetUnknown(StringType),
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
			elementType: StringType,
			elements:    []bool{true},
			expected:    SetUnknown(StringType),
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

			got, diags := SetValueFrom(context.Background(), testCase.elementType, testCase.elements)

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
func TestSetValue_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	knownSet := SetValueMust(StringType, []attr.Value{StringValue("test")})

	knownSet.Null = true

	if knownSet.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	knownSet.Unknown = true

	if knownSet.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	knownSet.Elems = []attr.Value{StringValue("not-test")}

	if knownSet.Elements()[0].Equal(StringValue("not-test")) {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestSetNull_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	nullSet := SetNull(StringType)

	nullSet.Null = false

	if !nullSet.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	nullSet.Unknown = true

	if nullSet.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	nullSet.Elems = []attr.Value{StringValue("test")}

	if len(nullSet.Elements()) > 0 {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestSetUnknown_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	unknownSet := SetUnknown(StringType)

	unknownSet.Null = true

	if unknownSet.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	unknownSet.Unknown = false

	if !unknownSet.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	unknownSet.Elems = []attr.Value{StringValue("test")}

	if len(unknownSet.Elements()) > 0 {
		t.Error("unexpected value update after Value field setting")
	}
}

func TestSetToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Set
		expectation tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"known": {
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"known-duplicates": {
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("hello"),
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
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringUnknown(),
					StringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"known-partial-null": {
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringNull(),
					StringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"unknown": {
			input:       SetUnknown(StringType),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"null": {
			input:       SetNull(StringType),
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
		},
		"deprecated-known": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"deprecated-known-duplicates": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "hello"},
				},
			},
			// Duplicate validation does not occur during this method.
			// This is okay, as tftypes allows duplicates.
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "hello"),
			}),
		},
		"deprecated-unknown": {
			input:       Set{ElemType: StringType, Unknown: true},
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       Set{ElemType: StringType, Null: true},
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
		},
		"deprecated-known-partial-unknown": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Unknown: true},
					String{Value: "hello, world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"deprecated-known-partial-null": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Null: true},
					String{Value: "hello, world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, nil),
				tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"no-elem-type": {
			input: Set{
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: tftypes.Value{},
			expectedErr: "cannot convert Set to tftypes.Value if ElemType field is not set",
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

func TestSetElements(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Set
		expected []attr.Value
	}{
		"known": {
			input:    SetValueMust(StringType, []attr.Value{StringValue("test")}),
			expected: []attr.Value{StringValue("test")},
		},
		"deprecated-known": {
			input:    Set{ElemType: StringType, Elems: []attr.Value{StringValue("test")}},
			expected: []attr.Value{StringValue("test")},
		},
		"null": {
			input:    SetNull(StringType),
			expected: nil,
		},
		"deprecated-null": {
			input:    Set{ElemType: StringType, Null: true},
			expected: nil,
		},
		"unknown": {
			input:    SetUnknown(StringType),
			expected: nil,
		},
		"deprecated-unknown": {
			input:    Set{ElemType: StringType, Unknown: true},
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

func TestSetElementType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Set
		expected attr.Type
	}{
		"known": {
			input:    SetValueMust(StringType, []attr.Value{StringValue("test")}),
			expected: StringType,
		},
		"deprecated-known": {
			input:    Set{ElemType: StringType, Elems: []attr.Value{StringValue("test")}},
			expected: StringType,
		},
		"null": {
			input:    SetNull(StringType),
			expected: StringType,
		},
		"deprecated-null": {
			input:    Set{ElemType: StringType, Null: true},
			expected: StringType,
		},
		"unknown": {
			input:    SetUnknown(StringType),
			expected: StringType,
		},
		"deprecated-unknown": {
			input:    Set{ElemType: StringType, Unknown: true},
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

func TestSetEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver Set
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: SetValueMust(
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
			expected: true,
		},
		"known-known-diff-value": {
			receiver: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("goodnight"),
					StringValue("moon"),
				},
			),
			expected: false,
		},
		"known-known-diff-length": {
			receiver: SetValueMust(
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
					StringValue("extra"),
				},
			),
			expected: false,
		},
		"known-known-diff-type": {
			receiver: SetValueMust(
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
			receiver: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringUnknown(),
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
		"known-known-diff-null": {
			receiver: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringNull(),
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
		"known-unknown": {
			receiver: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    SetUnknown(StringType),
			expected: false,
		},
		"known-null": {
			receiver: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    SetNull(StringType),
			expected: false,
		},
		"known-diff-type": {
			receiver: SetValueMust(
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
			expected: false,
		},
		"known-nil": {
			receiver: SetValueMust(
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
			receiver: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false, // intentional
		},
		"known-deprecated-unknown": {
			receiver: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    Set{ElemType: StringType, Unknown: true},
			expected: false,
		},
		"known-deprecated-null": {
			receiver: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			input:    Set{ElemType: StringType, Null: true},
			expected: false,
		},
		"deprecated-known-deprecated-known": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: true,
		},
		"deprecated-known-deprecated-known-diff-value": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "goodnight"},
					String{Value: "moon"},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-length": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: Set{
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
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: Set{
				ElemType: BoolType,
				Elems: []attr.Value{
					Bool{Value: false},
					Bool{Value: true},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-unknown": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Unknown: true},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-null": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Null: true},
				},
			},
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-unknown": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    Set{Unknown: true},
			expected: false,
		},
		"deprecated-known-deprecated-null": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    Set{Null: true},
			expected: false,
		},
		"deprecated-known-known": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expected: false, // intentional
		},
		"deprecated-known-unknown": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    SetUnknown(StringType),
			expected: false,
		},
		"deprecated-known-null": {
			receiver: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			input:    SetNull(StringType),
			expected: false,
		},
		"deprecated-known-diff-type": {
			receiver: Set{
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
			receiver: Set{
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

func TestSetIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Set
		expected bool
	}{
		"known": {
			input:    SetValueMust(StringType, []attr.Value{StringValue("test")}),
			expected: false,
		},
		"deprecated-known": {
			input:    Set{ElemType: StringType, Elems: []attr.Value{StringValue("test")}},
			expected: false,
		},
		"null": {
			input:    SetNull(StringType),
			expected: true,
		},
		"deprecated-null": {
			input:    Set{ElemType: StringType, Null: true},
			expected: true,
		},
		"unknown": {
			input:    SetUnknown(StringType),
			expected: false,
		},
		"deprecated-unknown": {
			input:    Set{ElemType: StringType, Unknown: true},
			expected: false,
		},
		"deprecated-invalid": {
			input:    Set{ElemType: StringType, Null: true, Unknown: true},
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

func TestSetIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Set
		expected bool
	}{
		"known": {
			input:    SetValueMust(StringType, []attr.Value{StringValue("test")}),
			expected: false,
		},
		"deprecated-known": {
			input:    Set{ElemType: StringType, Elems: []attr.Value{StringValue("test")}},
			expected: false,
		},
		"null": {
			input:    SetNull(StringType),
			expected: false,
		},
		"deprecated-null": {
			input:    Set{ElemType: StringType, Null: true},
			expected: false,
		},
		"unknown": {
			input:    SetUnknown(StringType),
			expected: true,
		},
		"deprecated-unknown": {
			input:    Set{ElemType: StringType, Unknown: true},
			expected: true,
		},
		"deprecated-invalid": {
			input:    Set{ElemType: StringType, Null: true, Unknown: true},
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

func TestSetString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Set
		expectation string
	}
	tests := map[string]testCase{
		"known": {
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expectation: `["hello","world"]`,
		},
		"known-set-of-sets": {
			input: SetValueMust(
				SetType{
					ElemType: StringType,
				},
				[]attr.Value{
					SetValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
					SetValueMust(
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
			input:       SetUnknown(StringType),
			expectation: "<unknown>",
		},
		"null": {
			input:       SetNull(StringType),
			expectation: "<null>",
		},
		"deprecated-known": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: `["hello","world"]`,
		},
		"deprecated-known-set-of-sets": {
			input: Set{
				ElemType: SetType{
					ElemType: StringType,
				},
				Elems: []attr.Value{
					Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					Set{
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
			input:       Set{Unknown: true},
			expectation: "<unknown>",
		},
		"deprecated-null": {
			input:       Set{Null: true},
			expectation: "<null>",
		},
		"default-empty": {
			input:       Set{},
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

func TestSetType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Set
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expectation: SetType{ElemType: StringType},
		},
		"known-set-of-sets": {
			input: SetValueMust(
				SetType{
					ElemType: StringType,
				},
				[]attr.Value{
					SetValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
					SetValueMust(
						StringType,
						[]attr.Value{
							StringValue("foo"),
							StringValue("bar"),
						},
					),
				},
			),
			expectation: SetType{
				ElemType: SetType{
					ElemType: StringType,
				},
			},
		},
		"unknown": {
			input:       SetUnknown(StringType),
			expectation: SetType{ElemType: StringType},
		},
		"null": {
			input:       SetNull(StringType),
			expectation: SetType{ElemType: StringType},
		},
		"deprecated-known": {
			input: Set{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "hello"},
					String{Value: "world"},
				},
			},
			expectation: SetType{ElemType: StringType},
		},
		"deprecated-known-set-of-sets": {
			input: Set{
				ElemType: SetType{
					ElemType: StringType,
				},
				Elems: []attr.Value{
					Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "foo"},
							String{Value: "bar"},
						},
					},
				},
			},
			expectation: SetType{
				ElemType: SetType{
					ElemType: StringType,
				},
			},
		},
		"deprecated-unknown": {
			input:       Set{ElemType: StringType, Unknown: true},
			expectation: SetType{ElemType: StringType},
		},
		"deprecated-null": {
			input:       Set{ElemType: StringType, Null: true},
			expectation: SetType{ElemType: StringType},
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
