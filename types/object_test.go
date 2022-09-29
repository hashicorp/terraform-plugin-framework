package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestObjectTypeTerraformType_simple(t *testing.T) {
	t.Parallel()
	result := ObjectType{
		AttrTypes: map[string]attr.Type{
			"foo": StringType,
			"bar": NumberType,
			"baz": BoolType,
		},
	}.TerraformType(context.Background())
	if diff := cmp.Diff(result, tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.Number,
			"baz": tftypes.Bool,
		},
	}); diff != "" {
		t.Errorf("unexpected result (+expected, -got): %s", diff)
	}
}

func TestObjectTypeTerraformType_empty(t *testing.T) {
	t.Parallel()
	result := ObjectType{}.TerraformType(context.Background())
	if diff := cmp.Diff(result, tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{},
	}); diff != "" {
		t.Errorf("unexpected result (+expected, -got): %s", diff)
	}
}

func TestObjectTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    ObjectType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"basic-object": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
					"b": BoolType,
					"c": NumberType,
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Bool,
					"c": tftypes.Number,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "red"),
				"b": tftypes.NewValue(tftypes.Bool, true),
				"c": tftypes.NewValue(tftypes.Number, 123),
			}),
			expected: Object{
				Attrs: map[string]attr.Value{
					"a": String{Value: "red"},
					"b": Bool{Value: true},
					"c": Number{Value: big.NewFloat(123)},
				},
				AttrTypes: map[string]attr.Type{
					"a": StringType,
					"b": BoolType,
					"c": NumberType,
				},
			},
		},
		"extra-attribute": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
					"b": tftypes.Bool,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "red"),
				"b": tftypes.NewValue(tftypes.Bool, true),
			}),
			expectedErr: `expected tftypes.Object["a":tftypes.String], got tftypes.Object["a":tftypes.String, "b":tftypes.Bool]`,
		},
		"missing-attribute": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
					"b": BoolType,
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.String, "red"),
			}),
			expectedErr: `expected tftypes.Object["a":tftypes.String, "b":tftypes.Bool], got tftypes.Object["a":tftypes.String]`,
		},
		"wrong-type": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
			},
			input:       tftypes.NewValue(tftypes.String, "hello"),
			expectedErr: `expected tftypes.Object["a":tftypes.String], got tftypes.String`,
		},
		"nil-type": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
			},
			input: tftypes.NewValue(nil, nil),
			expected: Object{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
				Null: true,
			},
		},
		"unknown": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, tftypes.UnknownValue),
			expected: Object{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
				Unknown: true,
			},
		},
		"null": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
			},
			input: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.String,
				},
			}, nil),
			expected: Object{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
				Null: true,
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := test.receiver.ValueFromTerraform(context.Background(), test.input)
			if err != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err.Error())
					return
				}
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, err.Error())
					return
				}
			}
			if test.expectedErr != "" && err == nil {
				t.Errorf("Expected err to be %q, got nil", test.expectedErr)
				return
			}
			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("unexpected result (-expected, +got): %s", diff)
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

func TestObjectTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver ObjectType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"c": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"c": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			expected: true,
		},
		"missing-attr": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"c": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			expected: false,
		},
		"extra-attr": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"c": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			expected: false,
		},
		"diff-attrs": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"e": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"c": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			expected: false,
		},
		"diff": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": BoolType,
				"c": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"c": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			expected: false,
		},
		"nested-diff": {
			receiver: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"c": BoolType,
				"d": ListType{
					ElemType: StringType,
				},
			}},
			input: ObjectType{AttrTypes: map[string]attr.Type{
				"a": StringType,
				"b": NumberType,
				"c": BoolType,
				"d": ListType{
					ElemType: BoolType,
				},
			}},
			expected: false,
		},
		"wrongType": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
			},
			input:    NumberType,
			expected: false,
		},
		"nil": {
			receiver: ObjectType{
				AttrTypes: map[string]attr.Type{
					"a": StringType,
				},
			},
			input:    nil,
			expected: false,
		},
		"nil-attrs": {
			receiver: ObjectType{},
			input:    ObjectType{},
			expected: true,
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

func TestObjectValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attributeTypes map[string]attr.Type
		attributes     map[string]attr.Value
		expected       Object
		expectedDiags  diag.Diagnostics
	}{
		"valid-no-attributes": {
			attributeTypes: map[string]attr.Type{},
			attributes:     map[string]attr.Value{},
			expected:       ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}),
		},
		"valid-attributes": {
			attributeTypes: map[string]attr.Type{
				"null":    StringType,
				"unknown": StringType,
				"known":   StringType,
			},
			attributes: map[string]attr.Value{
				"null":    StringNull(),
				"unknown": StringUnknown(),
				"known":   StringValue("test"),
			},
			expected: ObjectValueMust(
				map[string]attr.Type{
					"null":    StringType,
					"unknown": StringType,
					"known":   StringType,
				},
				map[string]attr.Value{
					"null":    StringNull(),
					"unknown": StringUnknown(),
					"known":   StringValue("test"),
				},
			),
		},
		"invalid-attribute-value": {
			attributeTypes: map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			},
			attributes: map[string]attr.Value{
				"string": StringValue("test"),
				"bool":   StringValue("test"),
			},
			expected: ObjectUnknown(map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Object Attribute Type",
					"While creating a Object value, an invalid attribute value was detected. "+
						"A Object must use a matching attribute type for the value. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Object Attribute Name (bool) Expected Type: types.BoolType\n"+
						"Object Attribute Name (bool) Given Type: types.StringType",
				),
			},
		},
		"invalid-extra-attribute": {
			attributeTypes: map[string]attr.Type{
				"string": StringType,
			},
			attributes: map[string]attr.Value{
				"string": StringValue("test"),
				"bool":   BoolValue(true),
			},
			expected: ObjectUnknown(map[string]attr.Type{
				"string": StringType,
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Extra Object Attribute Value",
					"While creating a Object value, an extra attribute value was detected. "+
						"A Object must not contain values beyond the expected attribute types. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Extra Object Attribute Name: bool",
				),
			},
		},
		"invalid-missing-attribute": {
			attributeTypes: map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			},
			attributes: map[string]attr.Value{
				"string": StringValue("test"),
			},
			expected: ObjectUnknown(map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing Object Attribute Value",
					"While creating a Object value, a missing attribute value was detected. "+
						"A Object must contain values for all attributes, even if null or unknown. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Object Attribute Name (bool) Expected Type: types.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := ObjectValue(testCase.attributeTypes, testCase.attributes)

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
func TestObjectValue_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	knownObject := ObjectValueMust(
		map[string]attr.Type{"test_attr": StringType},
		map[string]attr.Value{"test_attr": StringValue("test-value")},
	)

	knownObject.Null = true

	if knownObject.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	knownObject.Unknown = true

	if knownObject.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	knownObject.Attrs = map[string]attr.Value{"test_attr": StringValue("not-test-value")}

	if knownObject.Attributes()["test_attr"].Equal(StringValue("not-test-value")) {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestObjectNull_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	nullObject := ObjectNull(map[string]attr.Type{"test_attr": StringType})

	nullObject.Null = false

	if !nullObject.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	nullObject.Unknown = true

	if nullObject.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	nullObject.Attrs = map[string]attr.Value{"test_attr": StringValue("test")}

	if len(nullObject.Attributes()) > 0 {
		t.Error("unexpected value update after Value field setting")
	}
}

// This test verifies the assumptions that creating the Value via function then
// setting the fields directly has no effects.
func TestObjectUnknown_DeprecatedFieldSetting(t *testing.T) {
	t.Parallel()

	unknownObject := ObjectUnknown(map[string]attr.Type{"test_attr": StringType})

	unknownObject.Null = true

	if unknownObject.IsNull() {
		t.Error("unexpected null update after Null field setting")
	}

	unknownObject.Unknown = false

	if !unknownObject.IsUnknown() {
		t.Error("unexpected unknown update after Unknown field setting")
	}

	unknownObject.Attrs = map[string]attr.Value{"test_attr": StringValue("test")}

	if len(unknownObject.Attributes()) > 0 {
		t.Error("unexpected value update after Value field setting")
	}
}

func TestObjectAs_struct(t *testing.T) {
	t.Parallel()

	type myEmbeddedStruct struct {
		Red    string `tfsdk:"red"`
		Blue   List   `tfsdk:"blue"`
		Green  Number `tfsdk:"green"`
		Yellow int    `tfsdk:"yellow"`
	}
	type myStruct struct {
		A string           `tfsdk:"a"`
		B Bool             `tfsdk:"b"`
		C List             `tfsdk:"c"`
		D []string         `tfsdk:"d"`
		E []Bool           `tfsdk:"e"`
		F []List           `tfsdk:"f"`
		G Object           `tfsdk:"g"`
		H myEmbeddedStruct `tfsdk:"h"`
		I Object           `tfsdk:"i"`
	}
	object := Object{
		AttrTypes: map[string]attr.Type{
			"a": StringType,
			"b": BoolType,
			"c": ListType{ElemType: StringType},
			"d": ListType{ElemType: StringType},
			"e": ListType{ElemType: BoolType},
			"f": ListType{ElemType: ListType{ElemType: StringType}},
			"g": ObjectType{
				AttrTypes: map[string]attr.Type{
					"dogs":  NumberType,
					"cats":  NumberType,
					"names": ListType{ElemType: StringType},
				},
			},
			"h": ObjectType{
				AttrTypes: map[string]attr.Type{
					"red":    StringType,
					"blue":   ListType{ElemType: NumberType},
					"green":  NumberType,
					"yellow": NumberType,
				},
			},
			"i": ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":     StringType,
					"age":      NumberType,
					"opted_in": BoolType,
				},
			},
		},
		Attrs: map[string]attr.Value{
			"a": String{Value: "hello"},
			"b": Bool{Value: true},
			"c": List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "into"},
					String{Value: "the"},
					String{Unknown: true},
					String{Null: true},
				},
			},
			"d": List{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "it's"},
					String{Value: "getting"},
					String{Value: "hard"},
					String{Value: "to"},
					String{Value: "come"},
					String{Value: "up"},
					String{Value: "with"},
					String{Value: "test"},
					String{Value: "values"},
				},
			},
			"e": List{
				ElemType: BoolType,
				Elems: []attr.Value{
					Bool{Value: true},
					Bool{Value: false},
					Bool{Value: false},
					Bool{Value: true},
				},
			},
			"f": List{
				ElemType: ListType{
					ElemType: StringType,
				},
				Elems: []attr.Value{
					List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "head"},
							String{Value: "empty"},
						},
					},
					List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "no"},
							String{Value: "thoughts"},
						},
					},
				},
			},
			"g": Object{
				AttrTypes: map[string]attr.Type{
					"dogs":  NumberType,
					"cats":  NumberType,
					"names": ListType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"dogs": Number{Value: big.NewFloat(3)},
					"cats": Number{Value: big.NewFloat(5)},
					"names": List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "Roxy"},
							String{Value: "Jpeg"},
							String{Value: "Kupo"},
							String{Value: "Clawde"},
							String{Value: "Yeti"},
							String{Value: "Abby"},
							String{Value: "Ellie"},
							String{Value: "Lexi"},
						},
					},
				},
			},
			"h": Object{
				AttrTypes: map[string]attr.Type{
					"red":    StringType,
					"blue":   ListType{ElemType: NumberType},
					"green":  NumberType,
					"yellow": NumberType,
				},
				Attrs: map[string]attr.Value{
					"red": String{Value: "judge me not too harshly, future maintainers, this much random data is hard to come up with without getting weird."},
					"blue": List{
						ElemType: NumberType,
						Elems: []attr.Value{
							Number{Value: big.NewFloat(1)},
							Number{Value: big.NewFloat(2)},
							Number{Value: big.NewFloat(3)},
						},
					},
					"green":  Number{Value: big.NewFloat(123.456)},
					"yellow": Number{Value: big.NewFloat(123)},
				},
			},
			"i": Object{
				AttrTypes: map[string]attr.Type{
					"name":     StringType,
					"age":      NumberType,
					"opted_in": BoolType,
				},
				Attrs: map[string]attr.Value{
					"name":     String{Value: "J Doe"},
					"age":      Number{Value: big.NewFloat(28)},
					"opted_in": Bool{Value: true},
				},
			},
		},
	}
	var target myStruct
	diags := object.As(context.Background(), &target, ObjectAsOptions{})
	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags)
	}
	expected := myStruct{
		A: "hello",
		B: Bool{Value: true},
		C: List{
			ElemType: StringType,
			Elems: []attr.Value{
				String{Value: "into"},
				String{Value: "the"},
				String{Unknown: true},
				String{Null: true},
			},
		},
		D: []string{"it's", "getting", "hard", "to", "come", "up", "with", "test", "values"},
		E: []Bool{
			{Value: true},
			{Value: false},
			{Value: false},
			{Value: true},
		},
		F: []List{
			{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "head"},
					String{Value: "empty"},
				},
			},
			{
				ElemType: StringType,
				Elems: []attr.Value{
					String{Value: "no"},
					String{Value: "thoughts"},
				},
			},
		},
		G: Object{
			AttrTypes: map[string]attr.Type{
				"dogs":  NumberType,
				"cats":  NumberType,
				"names": ListType{ElemType: StringType},
			},
			Attrs: map[string]attr.Value{
				"dogs": Number{Value: big.NewFloat(3)},
				"cats": Number{Value: big.NewFloat(5)},
				"names": List{
					ElemType: StringType,
					Elems: []attr.Value{
						String{Value: "Roxy"},
						String{Value: "Jpeg"},
						String{Value: "Kupo"},
						String{Value: "Clawde"},
						String{Value: "Yeti"},
						String{Value: "Abby"},
						String{Value: "Ellie"},
						String{Value: "Lexi"},
					},
				},
			},
		},
		H: myEmbeddedStruct{
			Red: "judge me not too harshly, future maintainers, this much random data is hard to come up with without getting weird.",
			Blue: List{
				ElemType: NumberType,
				Elems: []attr.Value{
					Number{Value: big.NewFloat(1)},
					Number{Value: big.NewFloat(2)},
					Number{Value: big.NewFloat(3)},
				},
			},
			Green:  Number{Value: big.NewFloat(123.456)},
			Yellow: 123,
		},
		I: Object{
			AttrTypes: map[string]attr.Type{
				"name":     StringType,
				"age":      NumberType,
				"opted_in": BoolType,
			},
			Attrs: map[string]attr.Value{
				"name":     String{Value: "J Doe"},
				"age":      Number{Value: big.NewFloat(28)},
				"opted_in": Bool{Value: true},
			},
		},
	}
	if diff := cmp.Diff(expected, target); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestObjectAttributes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Object
		expected map[string]attr.Value
	}{
		"known": {
			input: ObjectValueMust(
				map[string]attr.Type{"test_attr": StringType},
				map[string]attr.Value{"test_attr": StringValue("test-value")},
			),
			expected: map[string]attr.Value{"test_attr": StringValue("test-value")},
		},
		"deprecated-known": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Attrs:     map[string]attr.Value{"test_attr": StringValue("test-value")},
			},
			expected: map[string]attr.Value{"test_attr": StringValue("test-value")},
		},
		"null": {
			input:    ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expected: nil,
		},
		"deprecated-null": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Null:      true,
			},
			expected: nil,
		},
		"unknown": {
			input:    ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
			expected: nil,
		},
		"deprecated-unknown": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Unknown:   true,
			},
			expected: nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.Attributes()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestObjectAttributeTypes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Object
		expected map[string]attr.Type
	}{
		"known": {
			input: ObjectValueMust(
				map[string]attr.Type{"test_attr": StringType},
				map[string]attr.Value{"test_attr": StringValue("test-value")},
			),
			expected: map[string]attr.Type{"test_attr": StringType},
		},
		"deprecated-known": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Attrs:     map[string]attr.Value{"test_attr": StringValue("test-value")},
			},
			expected: map[string]attr.Type{"test_attr": StringType},
		},
		"null": {
			input:    ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expected: map[string]attr.Type{"test_attr": StringType},
		},
		"deprecated-null": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Null:      true,
			},
			expected: map[string]attr.Type{"test_attr": StringType},
		},
		"unknown": {
			input:    ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
			expected: map[string]attr.Type{"test_attr": StringType},
		},
		"deprecated-unknown": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Unknown:   true,
			},
			expected: map[string]attr.Type{"test_attr": StringType},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.AttributeTypes(context.Background())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestObjectToTerraformValue(t *testing.T) {
	t.Parallel()
	type testCase struct {
		receiver    Object
		expected    tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"a": ListType{ElemType: StringType},
					"b": StringType,
					"c": BoolType,
					"d": NumberType,
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
					},
					"f": SetType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"a": List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					"b": String{Value: "woohoo"},
					"c": Bool{Value: true},
					"d": Number{Value: big.NewFloat(1234)},
					"e": Object{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
						Attrs: map[string]attr.Value{
							"name": String{Value: "testing123"},
						},
					},
					"f": Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
				},
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.List{ElementType: tftypes.String},
					"b": tftypes.String,
					"c": tftypes.Bool,
					"d": tftypes.Number,
					"e": tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}},
					"f": tftypes.Set{ElementType: tftypes.String},
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"b": tftypes.NewValue(tftypes.String, "woohoo"),
				"c": tftypes.NewValue(tftypes.Bool, true),
				"d": tftypes.NewValue(tftypes.Number, big.NewFloat(1234)),
				"e": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "testing123"),
				}),
				"f": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
			}),
		},
		"unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"a": ListType{ElemType: StringType},
					"b": StringType,
					"c": BoolType,
					"d": NumberType,
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
					},
					"f": SetType{ElemType: StringType},
				},
				Unknown: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.List{ElementType: tftypes.String},
					"b": tftypes.String,
					"c": tftypes.Bool,
					"d": tftypes.Number,
					"e": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name": tftypes.String,
						},
					},
					"f": tftypes.Set{ElementType: tftypes.String},
				},
			}, tftypes.UnknownValue),
		},
		"null": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"a": ListType{ElemType: StringType},
					"b": StringType,
					"c": BoolType,
					"d": NumberType,
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
					},
					"f": SetType{ElemType: StringType},
				},
				Null: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.List{ElementType: tftypes.String},
					"b": tftypes.String,
					"c": tftypes.Bool,
					"d": tftypes.Number,
					"e": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name": tftypes.String,
						},
					},
					"f": tftypes.Set{ElementType: tftypes.String},
				},
			}, nil),
		},
		"partial-unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"a": ListType{ElemType: StringType},
					"b": StringType,
					"c": BoolType,
					"d": NumberType,
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
					},
					"f": SetType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"a": List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					"b": String{Unknown: true},
					"c": Bool{Value: true},
					"d": Number{Value: big.NewFloat(1234)},
					"e": Object{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
						Attrs: map[string]attr.Value{
							"name": String{Value: "testing123"},
						},
					},
					"f": Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
				},
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.List{ElementType: tftypes.String},
					"b": tftypes.String,
					"c": tftypes.Bool,
					"d": tftypes.Number,
					"e": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name": tftypes.String,
						},
					},
					"f": tftypes.Set{ElementType: tftypes.String},
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"b": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"c": tftypes.NewValue(tftypes.Bool, true),
				"d": tftypes.NewValue(tftypes.Number, big.NewFloat(1234)),
				"e": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "testing123"),
				}),
				"f": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
			}),
		},
		"partial-null": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"a": ListType{ElemType: StringType},
					"b": StringType,
					"c": BoolType,
					"d": NumberType,
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
					},
					"f": SetType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"a": List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					"b": String{Null: true},
					"c": Bool{Value: true},
					"d": Number{Value: big.NewFloat(1234)},
					"e": Object{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
						Attrs: map[string]attr.Value{
							"name": String{Value: "testing123"},
						},
					},
					"f": Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
				},
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.List{ElementType: tftypes.String},
					"b": tftypes.String,
					"c": tftypes.Bool,
					"d": tftypes.Number,
					"e": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name": tftypes.String,
						},
					},
					"f": tftypes.Set{ElementType: tftypes.String},
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"b": tftypes.NewValue(tftypes.String, nil),
				"c": tftypes.NewValue(tftypes.Bool, true),
				"d": tftypes.NewValue(tftypes.Number, big.NewFloat(1234)),
				"e": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "testing123"),
				}),
				"f": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
			}),
		},
		"deep-partial-unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"a": ListType{ElemType: StringType},
					"b": StringType,
					"c": BoolType,
					"d": NumberType,
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
					},
					"f": SetType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"a": List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					"b": String{Value: "woohoo"},
					"c": Bool{Value: true},
					"d": Number{Value: big.NewFloat(1234)},
					"e": Object{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
						Attrs: map[string]attr.Value{
							"name": String{Unknown: true},
						},
					},
					"f": Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
				},
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.List{ElementType: tftypes.String},
					"b": tftypes.String,
					"c": tftypes.Bool,
					"d": tftypes.Number,
					"e": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name": tftypes.String,
						},
					},
					"f": tftypes.Set{ElementType: tftypes.String},
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"b": tftypes.NewValue(tftypes.String, "woohoo"),
				"c": tftypes.NewValue(tftypes.Bool, true),
				"d": tftypes.NewValue(tftypes.Number, big.NewFloat(1234)),
				"e": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				}),
				"f": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
			}),
		},
		"deep-partial-null": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"a": ListType{ElemType: StringType},
					"b": StringType,
					"c": BoolType,
					"d": NumberType,
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
					},
					"f": SetType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"a": List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					"b": String{Value: "woohoo"},
					"c": Bool{Value: true},
					"d": Number{Value: big.NewFloat(1234)},
					"e": Object{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
						Attrs: map[string]attr.Value{
							"name": String{Null: true},
						},
					},
					"f": Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
				},
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"a": tftypes.List{ElementType: tftypes.String},
					"b": tftypes.String,
					"c": tftypes.Bool,
					"d": tftypes.Number,
					"e": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name": tftypes.String,
						},
					},
					"f": tftypes.Set{ElementType: tftypes.String},
				},
			}, map[string]tftypes.Value{
				"a": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"b": tftypes.NewValue(tftypes.String, "woohoo"),
				"c": tftypes.NewValue(tftypes.Bool, true),
				"d": tftypes.NewValue(tftypes.Number, big.NewFloat(1234)),
				"e": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, nil),
				}),
				"f": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
			}),
		},
		"no-attr-types": {
			receiver: Object{
				Attrs: map[string]attr.Value{
					"a": List{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
					"b": String{Value: "woohoo"},
					"c": Bool{Value: true},
					"d": Number{Value: big.NewFloat(1234)},
					"e": Object{
						AttrTypes: map[string]attr.Type{
							"name": StringType,
						},
						Attrs: map[string]attr.Value{
							"name": String{Value: "testing123"},
						},
					},
					"f": Set{
						ElemType: StringType,
						Elems: []attr.Value{
							String{Value: "hello"},
							String{Value: "world"},
						},
					},
				},
			},
			expected:    tftypes.Value{},
			expectedErr: "cannot convert Object to tftypes.Value if AttrTypes field is not set",
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, gotErr := test.receiver.ToTerraformValue(context.Background())

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

			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestObjectEqual(t *testing.T) {
	t.Parallel()
	type testCase struct {
		receiver Object
		arg      attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("test"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			arg: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("test"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			expected: true,
		},
		"known-known-diff-value": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("test"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			arg: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("not-test"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			expected: false,
		},
		"known-known-diff-attribute-types": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"string": StringValue("hello"),
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"number": NumberValue(big.NewFloat(123)),
				},
			},
			expected: false,
		},
		"known-known-diff-unknown": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
				},
				map[string]attr.Value{
					"string": StringValue("hello"),
				},
			),
			arg: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
				},
				map[string]attr.Value{
					"string": StringUnknown(),
				},
			),
			expected: false,
		},
		"known-known-diff-null": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
				},
				map[string]attr.Value{
					"string": StringValue("hello"),
				},
			),
			arg: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
				},
				map[string]attr.Value{
					"string": StringNull(),
				},
			),
			expected: false,
		},
		"known-unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": StringValue("hello"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			},
			arg: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: false,
		},
		"known-null": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("hello"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			arg: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: false,
		},
		"known-diff-wrong-type": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("hello"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			arg:      StringValue("whoops"),
			expected: false,
		},
		"known-deprecated-known": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("hello"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			expected: false, // intentional
		},
		"known-deprecated-unknown": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("hello"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Unknown: true,
			},
			expected: false,
		},
		"known-deprecated-null": {
			receiver: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": StringValue("hello"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Null: true,
			},
			expected: false,
		},
		"unknown-known": {
			receiver: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			),
			expected: false,
		},
		"unknown-deprecated-known": {
			receiver: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			expected: false,
		},
		"unknown-unknown": {
			receiver: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: true,
		},
		"unknown-deprecated-unknown": {
			receiver: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Unknown: true,
			},
			expected: false, // intentional
		},
		"unknown-null": {
			receiver: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: false,
		},
		"unknown-deprecated-null": {
			receiver: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Null: true,
			},
			expected: false,
		},
		"null-known": {
			receiver: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			),
			expected: false,
		},
		"null-deprecated-known": {
			receiver: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			expected: false,
		},
		"null-unknown": {
			receiver: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: false,
		},
		"null-deprecated-unknown": {
			receiver: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Unknown: true,
			},
			expected: false,
		},
		"null-null": {
			receiver: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: true,
		},
		"null-deprecated-null": {
			receiver: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Null: true,
			},
			expected: false, // intentional
		},
		"deprecated-known-deprecated-known": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			expected: true,
		},
		"deprecated-known-deprecated-known-diff-value": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "world"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-attribute-types": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-attribute-count": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"list":   ListType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"list": List{ElemType: BoolType, Elems: []attr.Value{
						Bool{Value: true},
						Bool{Value: false},
					}},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Unknown: true},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-known-diff-null": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Null: true},
				},
			},
			expected: false,
		},
		"deprecated-known-deprecated-unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Unknown: true,
			},
			expected: false,
		},
		"deprecated-known-deprecated-null": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Null: true,
			},
			expected: false,
		},
		"deprecated-known-diff-wrong-type": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			arg:      String{Value: "whoops"},
			expected: false,
		},
		"deprecated-known-invalid-attribute-name": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
				},
				Attrs: map[string]attr.Value{
					"strng": String{Value: "hello"},
				},
			},
			expected: false,
		},
		"deprecated-known-known": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			arg: ObjectValueMust(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			),
			expected: false, // intentional
		},
		"deprecated-known-unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			arg: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: false,
		},
		"deprecated-known-null": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"bool":   Bool{Value: true},
					"number": Number{Value: big.NewFloat(123)},
				},
			},
			arg: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: false,
		},
		"deprecated-unknown-deprecated-unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Unknown: true,
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Unknown: true,
			},
			expected: true,
		},
		"deprecated-unknown-unknown": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Unknown: true,
			},
			arg: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: false, // intentional
		},
		"deprecated-null-deprecated-null": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Null: true,
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Null: true,
			},
			expected: true,
		},
		"deprecated-null-null": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
				Null: true,
			},
			arg: ObjectNull(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
					"number": NumberType,
				},
			),
			expected: false, // intentional
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.arg)
			if got != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestObjectIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Object
		expected bool
	}{
		"known": {
			input: ObjectValueMust(
				map[string]attr.Type{"test_attr": StringType},
				map[string]attr.Value{"test_attr": StringValue("test-value")},
			),
			expected: false,
		},
		"deprecated-known": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Attrs:     map[string]attr.Value{"test_attr": StringValue("test-value")},
			},
			expected: false,
		},
		"null": {
			input:    ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expected: true,
		},
		"deprecated-null": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Null:      true,
			},
			expected: true,
		},
		"unknown": {
			input:    ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
			expected: false,
		},
		"deprecated-unknown": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Unknown:   true,
			},
			expected: false,
		},
		"deprecated-invalid": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Null:      true,
				Unknown:   true,
			},
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

func TestObjectIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Object
		expected bool
	}{
		"known": {
			input: ObjectValueMust(
				map[string]attr.Type{"test_attr": StringType},
				map[string]attr.Value{"test_attr": StringValue("test-value")},
			),
			expected: false,
		},
		"deprecated-known": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Attrs:     map[string]attr.Value{"test_attr": StringValue("test-value")},
			},
			expected: false,
		},
		"null": {
			input:    ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expected: false,
		},
		"deprecated-null": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Null:      true,
			},
			expected: false,
		},
		"unknown": {
			input:    ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
			expected: true,
		},
		"deprecated-unknown": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Unknown:   true,
			},
			expected: true,
		},
		"deprecated-invalid": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Null:      true,
				Unknown:   true,
			},
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

func TestObjectString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Object
		expectation string
	}
	tests := map[string]testCase{
		"known": {
			input: ObjectValueMust(
				map[string]attr.Type{
					"alpha": StringType,
					"beta":  Int64Type,
					"gamma": Float64Type,
					"sigma": NumberType,
					"theta": BoolType,
				},
				map[string]attr.Value{
					"alpha": String{Value: "hello"},
					"beta":  Int64{Value: 98719827987189},
					"gamma": Float64{Value: -9876.782378},
					"sigma": Number{Unknown: true},
					"theta": Bool{Null: true},
				},
			),
			expectation: `{"alpha":"hello","beta":98719827987189,"gamma":-9876.782378,"sigma":<unknown>,"theta":<null>}`,
		},
		"known-object-of-objects": {
			input: ObjectValueMust(
				map[string]attr.Type{
					"alpha": ObjectType{
						AttrTypes: map[string]attr.Type{
							"one":   StringType,
							"two":   BoolType,
							"three": NumberType,
						},
					},
					"beta": ObjectType{
						AttrTypes: map[string]attr.Type{
							"uno": Int64Type,
							"due": BoolType,
							"tre": StringType,
						},
					},
					"gamma": Float64Type,
					"sigma": NumberType,
					"theta": BoolType,
				},
				map[string]attr.Value{
					"alpha": ObjectValueMust(
						map[string]attr.Type{
							"one":   StringType,
							"two":   BoolType,
							"three": NumberType,
						},
						map[string]attr.Value{
							"one":   String{Value: "1"},
							"two":   Bool{Value: true},
							"three": Number{Value: big.NewFloat(0.3)},
						},
					),
					"beta": ObjectValueMust(
						map[string]attr.Type{
							"uno": Int64Type,
							"due": BoolType,
							"tre": StringType,
						},
						map[string]attr.Value{
							"uno": Int64{Value: 1},
							"due": Bool{Value: false},
							"tre": String{Value: "3"},
						},
					),
					"gamma": Float64{Value: -9876.782378},
					"sigma": Number{Unknown: true},
					"theta": Bool{Null: true},
				},
			),
			expectation: `{"alpha":{"one":"1","three":0.3,"two":true},"beta":{"due":false,"tre":"3","uno":1},"gamma":-9876.782378,"sigma":<unknown>,"theta":<null>}`,
		},
		"unknown": {
			input:       ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
			expectation: "<unknown>",
		},
		"null": {
			input:       ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expectation: "<null>",
		},
		"deprecated-known": {
			input: Object{
				AttrTypes: map[string]attr.Type{
					"alpha": StringType,
					"beta":  Int64Type,
					"gamma": Float64Type,
					"sigma": NumberType,
					"theta": BoolType,
				},
				Attrs: map[string]attr.Value{
					"alpha": String{Value: "hello"},
					"beta":  Int64{Value: 98719827987189},
					"gamma": Float64{Value: -9876.782378},
					"sigma": Number{Unknown: true},
					"theta": Bool{Null: true},
				},
			},
			expectation: `{"alpha":"hello","beta":98719827987189,"gamma":-9876.782378,"sigma":<unknown>,"theta":<null>}`,
		},
		"deprecated-known-object-of-objects": {
			input: Object{
				AttrTypes: map[string]attr.Type{
					"alpha": ObjectType{
						AttrTypes: map[string]attr.Type{
							"one":   StringType,
							"two":   BoolType,
							"three": NumberType,
						},
					},
					"beta": ObjectType{
						AttrTypes: map[string]attr.Type{
							"uno": Int64Type,
							"due": BoolType,
							"tre": StringType,
						},
					},
					"gamma": Float64Type,
					"sigma": NumberType,
					"theta": BoolType,
				},
				Attrs: map[string]attr.Value{
					"alpha": Object{
						Attrs: map[string]attr.Value{
							"one":   String{Value: "1"},
							"two":   Bool{Value: true},
							"three": Number{Value: big.NewFloat(0.3)},
						},
					},
					"beta": Object{
						Attrs: map[string]attr.Value{
							"uno": Int64{Value: 1},
							"due": Bool{Value: false},
							"tre": String{Value: "3"},
						},
					},
					"gamma": Float64{Value: -9876.782378},
					"sigma": Number{Unknown: true},
					"theta": Bool{Null: true},
				},
			},
			expectation: `{"alpha":{"one":"1","three":0.3,"two":true},"beta":{"due":false,"tre":"3","uno":1},"gamma":-9876.782378,"sigma":<unknown>,"theta":<null>}`,
		},
		"deprecated-unknown": {
			input:       Object{Unknown: true},
			expectation: "<unknown>",
		},
		"deprecated-null": {
			input:       Object{Null: true},
			expectation: "<null>",
		},
		"default-empty": {
			input:       Object{},
			expectation: "{}",
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

func TestObjectType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Object
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: ObjectValueMust(
				map[string]attr.Type{
					"test_attr1": StringType,
					"test_attr2": StringType,
				},
				map[string]attr.Value{
					"test_attr1": StringValue("hello"),
					"test_attr2": StringValue("world"),
				},
			),
			expectation: ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr1": StringType,
					"test_attr2": StringType,
				},
			},
		},
		"known-object-of-objects": {
			input: ObjectValueMust(
				map[string]attr.Type{
					"test_attr1": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
					},
					"test_attr2": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
					},
				},
				map[string]attr.Value{
					"test_attr1": ObjectValueMust(
						map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
						map[string]attr.Value{
							"test_attr1": StringValue("hello"),
							"test_attr2": StringValue("world"),
						},
					),
					"test_attr2": ObjectValueMust(
						map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
						map[string]attr.Value{
							"test_attr1": StringValue("foo"),
							"test_attr2": StringValue("bar"),
						},
					),
				},
			),
			expectation: ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr1": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
					},
					"test_attr2": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
					},
				},
			},
		},
		"unknown": {
			input:       ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
			expectation: ObjectType{AttrTypes: map[string]attr.Type{"test_attr": StringType}},
		},
		"null": {
			input:       ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expectation: ObjectType{AttrTypes: map[string]attr.Type{"test_attr": StringType}},
		},
		"deprecated-known": {
			input: Object{
				AttrTypes: map[string]attr.Type{
					"test_attr1": StringType,
					"test_attr2": StringType,
				},
				Attrs: map[string]attr.Value{
					"test_attr1": String{Value: "hello"},
					"test_attr2": String{Value: "world"},
				},
			},
			expectation: ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr1": StringType,
					"test_attr2": StringType,
				},
			},
		},
		"deprecated-known-object-of-objects": {
			input: Object{
				AttrTypes: map[string]attr.Type{
					"test_attr1": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
					},
					"test_attr2": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
					},
				},
				Attrs: map[string]attr.Value{
					"test_attr1": ObjectValueMust(
						map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
						map[string]attr.Value{
							"test_attr1": StringValue("hello"),
							"test_attr2": StringValue("world"),
						},
					),
					"test_attr2": ObjectValueMust(
						map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
						map[string]attr.Value{
							"test_attr1": StringValue("foo"),
							"test_attr2": StringValue("bar"),
						},
					),
				},
			},
			expectation: ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr1": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
					},
					"test_attr2": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType,
							"test_attr2": StringType,
						},
					},
				},
			},
		},
		"deprecated-unknown": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Unknown:   true,
			},
			expectation: ObjectType{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
			},
		},
		"deprecated-null": {
			input: Object{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
				Null:      true,
			},
			expectation: ObjectType{
				AttrTypes: map[string]attr.Type{"test_attr": StringType},
			},
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
