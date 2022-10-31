package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
			expected: ObjectValueMust(
				map[string]attr.Type{
					"a": StringType,
					"b": BoolType,
					"c": NumberType,
				},
				map[string]attr.Value{
					"a": StringValue("red"),
					"b": BoolValue(true),
					"c": NumberValue(big.NewFloat(123)),
				},
			),
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
			expected: ObjectNull(
				map[string]attr.Type{
					"a": StringType,
				},
			),
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
			expected: ObjectUnknown(
				map[string]attr.Type{
					"a": StringType,
				},
			),
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
			expected: ObjectNull(
				map[string]attr.Type{
					"a": StringType,
				},
			),
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

func TestObjectValueFrom(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attributeTypes map[string]attr.Type
		attributes     any
		expected       Object
		expectedDiags  diag.Diagnostics
	}{
		"valid-*struct": {
			attributeTypes: map[string]attr.Type{
				"bool":   BoolType,
				"string": StringType,
			},
			attributes: pointer(struct {
				Bool   Bool   `tfsdk:"bool"`
				String String `tfsdk:"string"`
			}{
				Bool:   BoolValue(true),
				String: StringValue("test"),
			}),
			expected: ObjectValueMust(
				map[string]attr.Type{
					"bool":   BoolType,
					"string": StringType,
				},
				map[string]attr.Value{
					"bool":   BoolValue(true),
					"string": StringValue("test"),
				},
			),
		},
		"valid-struct": {
			attributeTypes: map[string]attr.Type{
				"bool":   BoolType,
				"string": StringType,
			},
			attributes: struct {
				Bool   Bool   `tfsdk:"bool"`
				String String `tfsdk:"string"`
			}{
				Bool:   BoolValue(true),
				String: StringValue("test"),
			},
			expected: ObjectValueMust(
				map[string]attr.Type{
					"bool":   BoolType,
					"string": StringType,
				},
				map[string]attr.Value{
					"bool":   BoolValue(true),
					"string": StringValue("test"),
				},
			),
		},
		"invalid-nil": {
			attributeTypes: map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			},
			attributes: nil,
			expected: ObjectUnknown(
				map[string]attr.Type{
					"string": StringType,
					"bool":   BoolType,
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"cannot construct attr.Type from <nil> (invalid)",
				),
			},
		},
		// This likely should be valid, however it is not currently.
		"invalid-map[string]attr.Value": {
			attributeTypes: map[string]attr.Type{
				"bool":   BoolType,
				"string": StringType,
			},
			attributes: map[string]attr.Value{
				"bool":   BoolNull(),
				"string": StringNull(),
			},
			expected: ObjectUnknown(
				map[string]attr.Type{
					"bool":   BoolType,
					"string": StringType,
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"cannot use type map[string]attr.Value as schema type types.ObjectType; types.ObjectType must be an attr.TypeWithElementType to hold map[string]attr.Value",
				),
			},
		},
		"invalid-not-struct": {
			attributeTypes: map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			},
			attributes: "oops",
			expected: ObjectUnknown(map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected tftypes.Object[\"bool\":tftypes.Bool, \"string\":tftypes.String], got tftypes.String",
				),
			},
		},
		"invalid-type": {
			attributeTypes: map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			},
			attributes: map[string]bool{"key1": true},
			expected: ObjectUnknown(map[string]attr.Type{
				"string": StringType,
				"bool":   BoolType,
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"cannot use type map[string]bool as schema type types.ObjectType; types.ObjectType must be an attr.TypeWithElementType to hold map[string]bool",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := ObjectValueFrom(context.Background(), testCase.attributeTypes, testCase.attributes)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				for _, d := range diags {
					t.Logf("%s\n%s", d.Summary(), d.Detail())
				}
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
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
	object := ObjectValueMust(
		map[string]attr.Type{
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
		map[string]attr.Value{
			"a": StringValue("hello"),
			"b": BoolValue(true),
			"c": ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("into"),
					StringValue("the"),
					StringUnknown(),
					StringNull(),
				},
			),
			"d": ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("it's"),
					StringValue("getting"),
					StringValue("hard"),
					StringValue("to"),
					StringValue("come"),
					StringValue("up"),
					StringValue("with"),
					StringValue("test"),
					StringValue("values"),
				},
			),
			"e": ListValueMust(
				BoolType,
				[]attr.Value{
					BoolValue(true),
					BoolValue(false),
					BoolValue(false),
					BoolValue(true),
				},
			),
			"f": ListValueMust(
				ListType{
					ElemType: StringType,
				},
				[]attr.Value{
					ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("head"),
							StringValue("empty"),
						},
					),
					ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("no"),
							StringValue("thoughts"),
						},
					),
				},
			),
			"g": ObjectValueMust(
				map[string]attr.Type{
					"dogs":  NumberType,
					"cats":  NumberType,
					"names": ListType{ElemType: StringType},
				},
				map[string]attr.Value{
					"dogs": NumberValue(big.NewFloat(3)),
					"cats": NumberValue(big.NewFloat(5)),
					"names": ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("Roxy"),
							StringValue("Jpeg"),
							StringValue("Kupo"),
							StringValue("Clawde"),
							StringValue("Yeti"),
							StringValue("Abby"),
							StringValue("Ellie"),
							StringValue("Lexi"),
						},
					),
				},
			),
			"h": ObjectValueMust(
				map[string]attr.Type{
					"red":    StringType,
					"blue":   ListType{ElemType: NumberType},
					"green":  NumberType,
					"yellow": NumberType,
				},
				map[string]attr.Value{
					"red": StringValue("judge me not too harshly, future maintainers, this much random data is hard to come up with without getting weird."),
					"blue": ListValueMust(
						NumberType,
						[]attr.Value{
							NumberValue(big.NewFloat(1)),
							NumberValue(big.NewFloat(2)),
							NumberValue(big.NewFloat(3)),
						},
					),
					"green":  NumberValue(big.NewFloat(123.456)),
					"yellow": NumberValue(big.NewFloat(123)),
				},
			),
			"i": ObjectValueMust(
				map[string]attr.Type{
					"name":     StringType,
					"age":      NumberType,
					"opted_in": BoolType,
				},
				map[string]attr.Value{
					"name":     StringValue("J Doe"),
					"age":      NumberValue(big.NewFloat(28)),
					"opted_in": BoolValue(true),
				},
			),
		},
	)
	var target myStruct
	diags := object.As(context.Background(), &target, ObjectAsOptions{})
	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags)
	}
	expected := myStruct{
		A: "hello",
		B: BoolValue(true),
		C: ListValueMust(
			StringType,
			[]attr.Value{
				StringValue("into"),
				StringValue("the"),
				StringUnknown(),
				StringNull(),
			},
		),
		D: []string{"it's", "getting", "hard", "to", "come", "up", "with", "test", "values"},
		E: []Bool{
			BoolValue(true),
			BoolValue(false),
			BoolValue(false),
			BoolValue(true),
		},
		F: []List{
			ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("head"),
					StringValue("empty"),
				},
			),
			ListValueMust(
				StringType,
				[]attr.Value{
					StringValue("no"),
					StringValue("thoughts"),
				},
			),
		},
		G: ObjectValueMust(
			map[string]attr.Type{
				"dogs":  NumberType,
				"cats":  NumberType,
				"names": ListType{ElemType: StringType},
			},
			map[string]attr.Value{
				"dogs": NumberValue(big.NewFloat(3)),
				"cats": NumberValue(big.NewFloat(5)),
				"names": ListValueMust(
					StringType,
					[]attr.Value{
						StringValue("Roxy"),
						StringValue("Jpeg"),
						StringValue("Kupo"),
						StringValue("Clawde"),
						StringValue("Yeti"),
						StringValue("Abby"),
						StringValue("Ellie"),
						StringValue("Lexi"),
					},
				),
			},
		),
		H: myEmbeddedStruct{
			Red: "judge me not too harshly, future maintainers, this much random data is hard to come up with without getting weird.",
			Blue: ListValueMust(
				NumberType,
				[]attr.Value{
					NumberValue(big.NewFloat(1)),
					NumberValue(big.NewFloat(2)),
					NumberValue(big.NewFloat(3)),
				},
			),
			Green:  NumberValue(big.NewFloat(123.456)),
			Yellow: 123,
		},
		I: ObjectValueMust(
			map[string]attr.Type{
				"name":     StringType,
				"age":      NumberType,
				"opted_in": BoolType,
			},
			map[string]attr.Value{
				"name":     StringValue("J Doe"),
				"age":      NumberValue(big.NewFloat(28)),
				"opted_in": BoolValue(true),
			},
		),
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
		"null": {
			input:    ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expected: nil,
		},
		"unknown": {
			input:    ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
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
		"null": {
			input:    ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expected: map[string]attr.Type{"test_attr": StringType},
		},
		"unknown": {
			input:    ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
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
			receiver: ObjectValueMust(
				map[string]attr.Type{
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
				map[string]attr.Value{
					"a": ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
					"b": StringValue("woohoo"),
					"c": BoolValue(true),
					"d": NumberValue(big.NewFloat(1234)),
					"e": ObjectValueMust(
						map[string]attr.Type{
							"name": StringType,
						},
						map[string]attr.Value{
							"name": StringValue("testing123"),
						},
					),
					"f": SetValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
				},
			),
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
			receiver: ObjectUnknown(
				map[string]attr.Type{
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
			),
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
			receiver: ObjectNull(
				map[string]attr.Type{
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
			),
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
			receiver: ObjectValueMust(
				map[string]attr.Type{
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
				map[string]attr.Value{
					"a": ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
					"b": StringUnknown(),
					"c": BoolValue(true),
					"d": NumberValue(big.NewFloat(1234)),
					"e": ObjectValueMust(
						map[string]attr.Type{
							"name": StringType,
						},
						map[string]attr.Value{
							"name": StringValue("testing123"),
						},
					),
					"f": SetValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
				},
			),
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
			receiver: ObjectValueMust(
				map[string]attr.Type{
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
				map[string]attr.Value{
					"a": ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
					"b": StringNull(),
					"c": BoolValue(true),
					"d": NumberValue(big.NewFloat(1234)),
					"e": ObjectValueMust(
						map[string]attr.Type{
							"name": StringType,
						},
						map[string]attr.Value{
							"name": StringValue("testing123"),
						},
					),
					"f": SetValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
				},
			),
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
			receiver: ObjectValueMust(
				map[string]attr.Type{
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
				map[string]attr.Value{
					"a": ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
					"b": StringValue("woohoo"),
					"c": BoolValue(true),
					"d": NumberValue(big.NewFloat(1234)),
					"e": ObjectValueMust(
						map[string]attr.Type{
							"name": StringType,
						},
						map[string]attr.Value{
							"name": StringUnknown(),
						},
					),
					"f": SetValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
				},
			),
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
			receiver: ObjectValueMust(
				map[string]attr.Type{
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
				map[string]attr.Value{
					"a": ListValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
					"b": StringValue("woohoo"),
					"c": BoolValue(true),
					"d": NumberValue(big.NewFloat(1234)),
					"e": ObjectValueMust(
						map[string]attr.Type{
							"name": StringType,
						},
						map[string]attr.Value{
							"name": StringNull(),
						},
					),
					"f": SetValueMust(
						StringType,
						[]attr.Value{
							StringValue("hello"),
							StringValue("world"),
						},
					),
				},
			),
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
					"number": NumberType,
				},
				map[string]attr.Value{
					"number": NumberValue(big.NewFloat(123)),
				},
			),
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
					"string": StringValue("hello"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
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
					"string": StringValue("hello"),
					"bool":   BoolValue(true),
					"number": NumberValue(big.NewFloat(123)),
				},
			),
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
		"null": {
			input:    ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expected: true,
		},
		"unknown": {
			input:    ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
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
		"null": {
			input:    ObjectNull(map[string]attr.Type{"test_attr": StringType}),
			expected: false,
		},
		"unknown": {
			input:    ObjectUnknown(map[string]attr.Type{"test_attr": StringType}),
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
					"alpha": StringValue("hello"),
					"beta":  Int64Value(98719827987189),
					"gamma": Float64Value(-9876.782378),
					"sigma": NumberUnknown(),
					"theta": BoolNull(),
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
							"one":   StringValue("1"),
							"two":   BoolValue(true),
							"three": NumberValue(big.NewFloat(0.3)),
						},
					),
					"beta": ObjectValueMust(
						map[string]attr.Type{
							"uno": Int64Type,
							"due": BoolType,
							"tre": StringType,
						},
						map[string]attr.Value{
							"uno": Int64Value(1),
							"due": BoolValue(false),
							"tre": StringValue("3"),
						},
					),
					"gamma": Float64Value(-9876.782378),
					"sigma": NumberUnknown(),
					"theta": BoolNull(),
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
		"zero-value": {
			input:       Object{},
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
