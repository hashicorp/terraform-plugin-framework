// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

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

func TestNewObjectValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attributeTypes map[string]attr.Type
		attributes     map[string]attr.Value
		expected       ObjectValue
		expectedDiags  diag.Diagnostics
	}{
		"valid-no-attributes": {
			attributeTypes: map[string]attr.Type{},
			attributes:     map[string]attr.Value{},
			expected:       NewObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{}),
		},
		"valid-attributes": {
			attributeTypes: map[string]attr.Type{
				"null":    StringType{},
				"unknown": StringType{},
				"known":   StringType{},
			},
			attributes: map[string]attr.Value{
				"null":    NewStringNull(),
				"unknown": NewStringUnknown(),
				"known":   NewStringValue("test"),
			},
			expected: NewObjectValueMust(
				map[string]attr.Type{
					"null":    StringType{},
					"unknown": StringType{},
					"known":   StringType{},
				},
				map[string]attr.Value{
					"null":    NewStringNull(),
					"unknown": NewStringUnknown(),
					"known":   NewStringValue("test"),
				},
			),
		},
		"invalid-attribute-value": {
			attributeTypes: map[string]attr.Type{
				"string": StringType{},
				"bool":   BoolType{},
			},
			attributes: map[string]attr.Value{
				"string": NewStringValue("test"),
				"bool":   NewStringValue("test"),
			},
			expected: NewObjectUnknown(map[string]attr.Type{
				"string": StringType{},
				"bool":   BoolType{},
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Object Attribute Type",
					"While creating a Object value, an invalid attribute value was detected. "+
						"A Object must use a matching attribute type for the value. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Object Attribute Name (bool) Expected Type: basetypes.BoolType\n"+
						"Object Attribute Name (bool) Given Type: basetypes.StringType",
				),
			},
		},
		"invalid-extra-attribute": {
			attributeTypes: map[string]attr.Type{
				"string": StringType{},
			},
			attributes: map[string]attr.Value{
				"string": NewStringValue("test"),
				"bool":   NewBoolValue(true),
			},
			expected: NewObjectUnknown(map[string]attr.Type{
				"string": StringType{},
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
				"string": StringType{},
				"bool":   BoolType{},
			},
			attributes: map[string]attr.Value{
				"string": NewStringValue("test"),
			},
			expected: NewObjectUnknown(map[string]attr.Type{
				"string": StringType{},
				"bool":   BoolType{},
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing Object Attribute Value",
					"While creating a Object value, a missing attribute value was detected. "+
						"A Object must contain values for all attributes, even if null or unknown. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Object Attribute Name (bool) Expected Type: basetypes.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := NewObjectValue(testCase.attributeTypes, testCase.attributes)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestNewObjectValueFrom(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attributeTypes map[string]attr.Type
		attributes     any
		expected       ObjectValue
		expectedDiags  diag.Diagnostics
	}{
		"valid-*struct": {
			attributeTypes: map[string]attr.Type{
				"bool":   BoolType{},
				"string": StringType{},
			},
			attributes: pointer(struct {
				Bool   BoolValue   `tfsdk:"bool"`
				String StringValue `tfsdk:"string"`
			}{
				Bool:   NewBoolValue(true),
				String: NewStringValue("test"),
			}),
			expected: NewObjectValueMust(
				map[string]attr.Type{
					"bool":   BoolType{},
					"string": StringType{},
				},
				map[string]attr.Value{
					"bool":   NewBoolValue(true),
					"string": NewStringValue("test"),
				},
			),
		},
		"valid-struct": {
			attributeTypes: map[string]attr.Type{
				"bool":   BoolType{},
				"string": StringType{},
			},
			attributes: struct {
				Bool   BoolValue   `tfsdk:"bool"`
				String StringValue `tfsdk:"string"`
			}{
				Bool:   NewBoolValue(true),
				String: NewStringValue("test"),
			},
			expected: NewObjectValueMust(
				map[string]attr.Type{
					"bool":   BoolType{},
					"string": StringType{},
				},
				map[string]attr.Value{
					"bool":   NewBoolValue(true),
					"string": NewStringValue("test"),
				},
			),
		},
		"invalid-nil": {
			attributeTypes: map[string]attr.Type{
				"string": StringType{},
				"bool":   BoolType{},
			},
			attributes: nil,
			expected: NewObjectUnknown(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
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
				"bool":   BoolType{},
				"string": StringType{},
			},
			attributes: map[string]attr.Value{
				"bool":   NewBoolNull(),
				"string": NewStringNull(),
			},
			expected: NewObjectUnknown(
				map[string]attr.Type{
					"bool":   BoolType{},
					"string": StringType{},
				},
			),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"cannot use type map[string]attr.Value as schema type basetypes.ObjectType; basetypes.ObjectType must be an attr.TypeWithElementType to hold map[string]attr.Value",
				),
			},
		},
		"invalid-not-struct": {
			attributeTypes: map[string]attr.Type{
				"string": StringType{},
				"bool":   BoolType{},
			},
			attributes: "oops",
			expected: NewObjectUnknown(map[string]attr.Type{
				"string": StringType{},
				"bool":   BoolType{},
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
				"string": StringType{},
				"bool":   BoolType{},
			},
			attributes: map[string]bool{"key1": true},
			expected: NewObjectUnknown(map[string]attr.Type{
				"string": StringType{},
				"bool":   BoolType{},
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"cannot use type map[string]bool as schema type basetypes.ObjectType; basetypes.ObjectType must be an attr.TypeWithElementType to hold map[string]bool",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := NewObjectValueFrom(context.Background(), testCase.attributeTypes, testCase.attributes)

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
		Red    string      `tfsdk:"red"`
		Blue   ListValue   `tfsdk:"blue"`
		Green  NumberValue `tfsdk:"green"`
		Yellow int         `tfsdk:"yellow"`
	}
	type myStruct struct {
		A string           `tfsdk:"a"`
		B BoolValue        `tfsdk:"b"`
		C ListValue        `tfsdk:"c"`
		D []string         `tfsdk:"d"`
		E []BoolValue      `tfsdk:"e"`
		F []ListValue      `tfsdk:"f"`
		G ObjectValue      `tfsdk:"g"`
		H myEmbeddedStruct `tfsdk:"h"`
		I ObjectValue      `tfsdk:"i"`
	}
	object := NewObjectValueMust(
		map[string]attr.Type{
			"a": StringType{},
			"b": BoolType{},
			"c": ListType{ElemType: StringType{}},
			"d": ListType{ElemType: StringType{}},
			"e": ListType{ElemType: BoolType{}},
			"f": ListType{ElemType: ListType{ElemType: StringType{}}},
			"g": ObjectType{
				AttrTypes: map[string]attr.Type{
					"dogs":  NumberType{},
					"cats":  NumberType{},
					"names": ListType{ElemType: StringType{}},
				},
			},
			"h": ObjectType{
				AttrTypes: map[string]attr.Type{
					"red":    StringType{},
					"blue":   ListType{ElemType: NumberType{}},
					"green":  NumberType{},
					"yellow": NumberType{},
				},
			},
			"i": ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":     StringType{},
					"age":      NumberType{},
					"opted_in": BoolType{},
				},
			},
		},
		map[string]attr.Value{
			"a": NewStringValue("hello"),
			"b": NewBoolValue(true),
			"c": NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("into"),
					NewStringValue("the"),
					NewStringUnknown(),
					NewStringNull(),
				},
			),
			"d": NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("it's"),
					NewStringValue("getting"),
					NewStringValue("hard"),
					NewStringValue("to"),
					NewStringValue("come"),
					NewStringValue("up"),
					NewStringValue("with"),
					NewStringValue("test"),
					NewStringValue("values"),
				},
			),
			"e": NewListValueMust(
				BoolType{},
				[]attr.Value{
					NewBoolValue(true),
					NewBoolValue(false),
					NewBoolValue(false),
					NewBoolValue(true),
				},
			),
			"f": NewListValueMust(
				ListType{
					ElemType: StringType{},
				},
				[]attr.Value{
					NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("head"),
							NewStringValue("empty"),
						},
					),
					NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("no"),
							NewStringValue("thoughts"),
						},
					),
				},
			),
			"g": NewObjectValueMust(
				map[string]attr.Type{
					"dogs":  NumberType{},
					"cats":  NumberType{},
					"names": ListType{ElemType: StringType{}},
				},
				map[string]attr.Value{
					"dogs": NewNumberValue(big.NewFloat(3)),
					"cats": NewNumberValue(big.NewFloat(5)),
					"names": NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("Roxy"),
							NewStringValue("Jpeg"),
							NewStringValue("Kupo"),
							NewStringValue("Clawde"),
							NewStringValue("Yeti"),
							NewStringValue("Abby"),
							NewStringValue("Ellie"),
							NewStringValue("Lexi"),
						},
					),
				},
			),
			"h": NewObjectValueMust(
				map[string]attr.Type{
					"red":    StringType{},
					"blue":   ListType{ElemType: NumberType{}},
					"green":  NumberType{},
					"yellow": NumberType{},
				},
				map[string]attr.Value{
					"red": NewStringValue("judge me not too harshly, future maintainers, this much random data is hard to come up with without getting weird."),
					"blue": NewListValueMust(
						NumberType{},
						[]attr.Value{
							NewNumberValue(big.NewFloat(1)),
							NewNumberValue(big.NewFloat(2)),
							NewNumberValue(big.NewFloat(3)),
						},
					),
					"green":  NewNumberValue(big.NewFloat(123.456)),
					"yellow": NewNumberValue(big.NewFloat(123)),
				},
			),
			"i": NewObjectValueMust(
				map[string]attr.Type{
					"name":     StringType{},
					"age":      NumberType{},
					"opted_in": BoolType{},
				},
				map[string]attr.Value{
					"name":     NewStringValue("J Doe"),
					"age":      NewNumberValue(big.NewFloat(28)),
					"opted_in": NewBoolValue(true),
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
		B: NewBoolValue(true),
		C: NewListValueMust(
			StringType{},
			[]attr.Value{
				NewStringValue("into"),
				NewStringValue("the"),
				NewStringUnknown(),
				NewStringNull(),
			},
		),
		D: []string{"it's", "getting", "hard", "to", "come", "up", "with", "test", "values"},
		E: []BoolValue{
			NewBoolValue(true),
			NewBoolValue(false),
			NewBoolValue(false),
			NewBoolValue(true),
		},
		F: []ListValue{
			NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("head"),
					NewStringValue("empty"),
				},
			),
			NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("no"),
					NewStringValue("thoughts"),
				},
			),
		},
		G: NewObjectValueMust(
			map[string]attr.Type{
				"dogs":  NumberType{},
				"cats":  NumberType{},
				"names": ListType{ElemType: StringType{}},
			},
			map[string]attr.Value{
				"dogs": NewNumberValue(big.NewFloat(3)),
				"cats": NewNumberValue(big.NewFloat(5)),
				"names": NewListValueMust(
					StringType{},
					[]attr.Value{
						NewStringValue("Roxy"),
						NewStringValue("Jpeg"),
						NewStringValue("Kupo"),
						NewStringValue("Clawde"),
						NewStringValue("Yeti"),
						NewStringValue("Abby"),
						NewStringValue("Ellie"),
						NewStringValue("Lexi"),
					},
				),
			},
		),
		H: myEmbeddedStruct{
			Red: "judge me not too harshly, future maintainers, this much random data is hard to come up with without getting weird.",
			Blue: NewListValueMust(
				NumberType{},
				[]attr.Value{
					NewNumberValue(big.NewFloat(1)),
					NewNumberValue(big.NewFloat(2)),
					NewNumberValue(big.NewFloat(3)),
				},
			),
			Green:  NewNumberValue(big.NewFloat(123.456)),
			Yellow: 123,
		},
		I: NewObjectValueMust(
			map[string]attr.Type{
				"name":     StringType{},
				"age":      NumberType{},
				"opted_in": BoolType{},
			},
			map[string]attr.Value{
				"name":     NewStringValue("J Doe"),
				"age":      NewNumberValue(big.NewFloat(28)),
				"opted_in": NewBoolValue(true),
			},
		),
	}
	if diff := cmp.Diff(expected, target); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestObjectValueAttributes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ObjectValue
		expected map[string]attr.Value
	}{
		"known": {
			input: NewObjectValueMust(
				map[string]attr.Type{"test_attr": StringType{}},
				map[string]attr.Value{"test_attr": NewStringValue("test-value")},
			),
			expected: map[string]attr.Value{"test_attr": NewStringValue("test-value")},
		},
		"null": {
			input:    NewObjectNull(map[string]attr.Type{"test_attr": StringType{}}),
			expected: map[string]attr.Value{},
		},
		"unknown": {
			input:    NewObjectUnknown(map[string]attr.Type{"test_attr": StringType{}}),
			expected: map[string]attr.Value{},
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

func TestObjectValueAttributes_immutable(t *testing.T) {
	t.Parallel()

	value := NewObjectValueMust(
		map[string]attr.Type{"test": StringType{}},
		map[string]attr.Value{"test": NewStringValue("original")},
	)
	expected := NewObjectValueMust(
		map[string]attr.Type{"test": StringType{}},
		map[string]attr.Value{"test": NewStringValue("original")},
	)
	value.Attributes()["test"] = NewStringValue("modified")

	if !value.Equal(expected) {
		t.Fatal("unexpected Attributes mutation")
	}
}

func TestObjectValueAttributeTypes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ObjectValue
		expected map[string]attr.Type
	}{
		"known": {
			input: NewObjectValueMust(
				map[string]attr.Type{"test_attr": StringType{}},
				map[string]attr.Value{"test_attr": NewStringValue("test-value")},
			),
			expected: map[string]attr.Type{"test_attr": StringType{}},
		},
		"null": {
			input:    NewObjectNull(map[string]attr.Type{"test_attr": StringType{}}),
			expected: map[string]attr.Type{"test_attr": StringType{}},
		},
		"unknown": {
			input:    NewObjectUnknown(map[string]attr.Type{"test_attr": StringType{}}),
			expected: map[string]attr.Type{"test_attr": StringType{}},
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

func TestObjectValueAttributeTypes_immutable(t *testing.T) {
	t.Parallel()

	value := NewObjectValueMust(
		map[string]attr.Type{"test": StringType{}},
		map[string]attr.Value{"test": NewStringValue("original")},
	)
	expected := NewObjectValueMust(
		map[string]attr.Type{"test": StringType{}},
		map[string]attr.Value{"test": NewStringValue("original")},
	)
	value.AttributeTypes(context.Background())["test"] = BoolType{}

	if !value.Equal(expected) {
		t.Fatal("unexpected AttributeTypes mutation")
	}
}

func TestObjectValueToTerraformValue(t *testing.T) {
	t.Parallel()
	type testCase struct {
		receiver    ObjectValue
		expected    tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"a": ListType{ElemType: StringType{}},
					"b": StringType{},
					"c": BoolType{},
					"d": NumberType{},
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType{},
						},
					},
					"f": SetType{ElemType: StringType{}},
				},
				map[string]attr.Value{
					"a": NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					"b": NewStringValue("woohoo"),
					"c": NewBoolValue(true),
					"d": NewNumberValue(big.NewFloat(1234)),
					"e": NewObjectValueMust(
						map[string]attr.Type{
							"name": StringType{},
						},
						map[string]attr.Value{
							"name": NewStringValue("testing123"),
						},
					),
					"f": NewSetValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
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
			receiver: NewObjectUnknown(
				map[string]attr.Type{
					"a": ListType{ElemType: StringType{}},
					"b": StringType{},
					"c": BoolType{},
					"d": NumberType{},
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType{},
						},
					},
					"f": SetType{ElemType: StringType{}},
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
			receiver: NewObjectNull(
				map[string]attr.Type{
					"a": ListType{ElemType: StringType{}},
					"b": StringType{},
					"c": BoolType{},
					"d": NumberType{},
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType{},
						},
					},
					"f": SetType{ElemType: StringType{}},
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
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"a": ListType{ElemType: StringType{}},
					"b": StringType{},
					"c": BoolType{},
					"d": NumberType{},
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType{},
						},
					},
					"f": SetType{ElemType: StringType{}},
				},
				map[string]attr.Value{
					"a": NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					"b": NewStringUnknown(),
					"c": NewBoolValue(true),
					"d": NewNumberValue(big.NewFloat(1234)),
					"e": NewObjectValueMust(
						map[string]attr.Type{
							"name": StringType{},
						},
						map[string]attr.Value{
							"name": NewStringValue("testing123"),
						},
					),
					"f": NewSetValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
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
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"a": ListType{ElemType: StringType{}},
					"b": StringType{},
					"c": BoolType{},
					"d": NumberType{},
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType{},
						},
					},
					"f": SetType{ElemType: StringType{}},
				},
				map[string]attr.Value{
					"a": NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					"b": NewStringNull(),
					"c": NewBoolValue(true),
					"d": NewNumberValue(big.NewFloat(1234)),
					"e": NewObjectValueMust(
						map[string]attr.Type{
							"name": StringType{},
						},
						map[string]attr.Value{
							"name": NewStringValue("testing123"),
						},
					),
					"f": NewSetValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
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
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"a": ListType{ElemType: StringType{}},
					"b": StringType{},
					"c": BoolType{},
					"d": NumberType{},
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType{},
						},
					},
					"f": SetType{ElemType: StringType{}},
				},
				map[string]attr.Value{
					"a": NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					"b": NewStringValue("woohoo"),
					"c": NewBoolValue(true),
					"d": NewNumberValue(big.NewFloat(1234)),
					"e": NewObjectValueMust(
						map[string]attr.Type{
							"name": StringType{},
						},
						map[string]attr.Value{
							"name": NewStringUnknown(),
						},
					),
					"f": NewSetValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
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
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"a": ListType{ElemType: StringType{}},
					"b": StringType{},
					"c": BoolType{},
					"d": NumberType{},
					"e": ObjectType{
						AttrTypes: map[string]attr.Type{
							"name": StringType{},
						},
					},
					"f": SetType{ElemType: StringType{}},
				},
				map[string]attr.Value{
					"a": NewListValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
						},
					),
					"b": NewStringValue("woohoo"),
					"c": NewBoolValue(true),
					"d": NewNumberValue(big.NewFloat(1234)),
					"e": NewObjectValueMust(
						map[string]attr.Type{
							"name": StringType{},
						},
						map[string]attr.Value{
							"name": NewStringNull(),
						},
					),
					"f": NewSetValueMust(
						StringType{},
						[]attr.Value{
							NewStringValue("hello"),
							NewStringValue("world"),
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

func TestObjectValueEqual(t *testing.T) {
	t.Parallel()
	type testCase struct {
		receiver ObjectValue
		arg      attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("test"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			arg: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("test"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			expected: true,
		},
		"known-known-diff-value": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("test"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			arg: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("not-test"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			expected: false,
		},
		"known-known-diff-attribute-types": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("hello"),
				},
			),
			arg: NewObjectValueMust(
				map[string]attr.Type{
					"number": NumberType{},
				},
				map[string]attr.Value{
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			expected: false,
		},
		"known-known-diff-unknown": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("hello"),
				},
			),
			arg: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
				},
				map[string]attr.Value{
					"string": NewStringUnknown(),
				},
			),
			expected: false,
		},
		"known-known-diff-null": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("hello"),
				},
			),
			arg: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
				},
				map[string]attr.Value{
					"string": NewStringNull(),
				},
			),
			expected: false,
		},
		"known-unknown": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("hello"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			arg: NewObjectUnknown(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			expected: false,
		},
		"known-null": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("hello"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			arg: NewObjectNull(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			expected: false,
		},
		"known-diff-wrong-type": {
			receiver: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("hello"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			arg:      NewStringValue("whoops"),
			expected: false,
		},
		"unknown-known": {
			receiver: NewObjectUnknown(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			arg: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("hello"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			expected: false,
		},
		"unknown-unknown": {
			receiver: NewObjectUnknown(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			arg: NewObjectUnknown(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			expected: true,
		},
		"unknown-null": {
			receiver: NewObjectUnknown(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			arg: NewObjectNull(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			expected: false,
		},
		"null-known": {
			receiver: NewObjectNull(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			arg: NewObjectValueMust(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
				map[string]attr.Value{
					"string": NewStringValue("hello"),
					"bool":   NewBoolValue(true),
					"number": NewNumberValue(big.NewFloat(123)),
				},
			),
			expected: false,
		},
		"null-unknown": {
			receiver: NewObjectNull(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			arg: NewObjectUnknown(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			expected: false,
		},
		"null-null": {
			receiver: NewObjectNull(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
				},
			),
			arg: NewObjectNull(
				map[string]attr.Type{
					"string": StringType{},
					"bool":   BoolType{},
					"number": NumberType{},
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

func TestObjectValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ObjectValue
		expected bool
	}{
		"known": {
			input: NewObjectValueMust(
				map[string]attr.Type{"test_attr": StringType{}},
				map[string]attr.Value{"test_attr": NewStringValue("test-value")},
			),
			expected: false,
		},
		"null": {
			input:    NewObjectNull(map[string]attr.Type{"test_attr": StringType{}}),
			expected: true,
		},
		"unknown": {
			input:    NewObjectUnknown(map[string]attr.Type{"test_attr": StringType{}}),
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

func TestObjectValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    ObjectValue
		expected bool
	}{
		"known": {
			input: NewObjectValueMust(
				map[string]attr.Type{"test_attr": StringType{}},
				map[string]attr.Value{"test_attr": NewStringValue("test-value")},
			),
			expected: false,
		},
		"null": {
			input:    NewObjectNull(map[string]attr.Type{"test_attr": StringType{}}),
			expected: false,
		},
		"unknown": {
			input:    NewObjectUnknown(map[string]attr.Type{"test_attr": StringType{}}),
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

func TestObjectValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       ObjectValue
		expectation string
	}
	tests := map[string]testCase{
		"known": {
			input: NewObjectValueMust(
				map[string]attr.Type{
					"alpha": StringType{},
					"beta":  Int64Type{},
					"gamma": Float64Type{},
					"sigma": NumberType{},
					"theta": BoolType{},
				},
				map[string]attr.Value{
					"alpha": NewStringValue("hello"),
					"beta":  NewInt64Value(98719827987189),
					"gamma": NewFloat64Value(-9876.782378),
					"sigma": NewNumberUnknown(),
					"theta": NewBoolNull(),
				},
			),
			expectation: `{"alpha":"hello","beta":98719827987189,"gamma":-9876.782378,"sigma":<unknown>,"theta":<null>}`,
		},
		"known-object-of-objects": {
			input: NewObjectValueMust(
				map[string]attr.Type{
					"alpha": ObjectType{
						AttrTypes: map[string]attr.Type{
							"one":   StringType{},
							"two":   BoolType{},
							"three": NumberType{},
						},
					},
					"beta": ObjectType{
						AttrTypes: map[string]attr.Type{
							"uno": Int64Type{},
							"due": BoolType{},
							"tre": StringType{},
						},
					},
					"gamma": Float64Type{},
					"sigma": NumberType{},
					"theta": BoolType{},
				},
				map[string]attr.Value{
					"alpha": NewObjectValueMust(
						map[string]attr.Type{
							"one":   StringType{},
							"two":   BoolType{},
							"three": NumberType{},
						},
						map[string]attr.Value{
							"one":   NewStringValue("1"),
							"two":   NewBoolValue(true),
							"three": NewNumberValue(big.NewFloat(0.3)),
						},
					),
					"beta": NewObjectValueMust(
						map[string]attr.Type{
							"uno": Int64Type{},
							"due": BoolType{},
							"tre": StringType{},
						},
						map[string]attr.Value{
							"uno": NewInt64Value(1),
							"due": NewBoolValue(false),
							"tre": NewStringValue("3"),
						},
					),
					"gamma": NewFloat64Value(-9876.782378),
					"sigma": NewNumberUnknown(),
					"theta": NewBoolNull(),
				},
			),
			expectation: `{"alpha":{"one":"1","three":0.3,"two":true},"beta":{"due":false,"tre":"3","uno":1},"gamma":-9876.782378,"sigma":<unknown>,"theta":<null>}`,
		},
		"unknown": {
			input:       NewObjectUnknown(map[string]attr.Type{"test_attr": StringType{}}),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewObjectNull(map[string]attr.Type{"test_attr": StringType{}}),
			expectation: "<null>",
		},
		"zero-value": {
			input:       ObjectValue{},
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

func TestObjectValueType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       ObjectValue
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: NewObjectValueMust(
				map[string]attr.Type{
					"test_attr1": StringType{},
					"test_attr2": StringType{},
				},
				map[string]attr.Value{
					"test_attr1": NewStringValue("hello"),
					"test_attr2": NewStringValue("world"),
				},
			),
			expectation: ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr1": StringType{},
					"test_attr2": StringType{},
				},
			},
		},
		"known-object-of-objects": {
			input: NewObjectValueMust(
				map[string]attr.Type{
					"test_attr1": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType{},
							"test_attr2": StringType{},
						},
					},
					"test_attr2": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType{},
							"test_attr2": StringType{},
						},
					},
				},
				map[string]attr.Value{
					"test_attr1": NewObjectValueMust(
						map[string]attr.Type{
							"test_attr1": StringType{},
							"test_attr2": StringType{},
						},
						map[string]attr.Value{
							"test_attr1": NewStringValue("hello"),
							"test_attr2": NewStringValue("world"),
						},
					),
					"test_attr2": NewObjectValueMust(
						map[string]attr.Type{
							"test_attr1": StringType{},
							"test_attr2": StringType{},
						},
						map[string]attr.Value{
							"test_attr1": NewStringValue("foo"),
							"test_attr2": NewStringValue("bar"),
						},
					),
				},
			),
			expectation: ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_attr1": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType{},
							"test_attr2": StringType{},
						},
					},
					"test_attr2": ObjectType{
						AttrTypes: map[string]attr.Type{
							"test_attr1": StringType{},
							"test_attr2": StringType{},
						},
					},
				},
			},
		},
		"unknown": {
			input:       NewObjectUnknown(map[string]attr.Type{"test_attr": StringType{}}),
			expectation: ObjectType{AttrTypes: map[string]attr.Type{"test_attr": StringType{}}},
		},
		"null": {
			input:       NewObjectNull(map[string]attr.Type{"test_attr": StringType{}}),
			expectation: ObjectType{AttrTypes: map[string]attr.Type{"test_attr": StringType{}}},
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
