package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := test.receiver.ToTerraformValue(context.Background())
			if err != nil {
				if test.expectedErr == "" {
					t.Errorf("unexpected error: %s", err)
					return
				}
				if test.expectedErr != err.Error() {
					t.Errorf("expected error to be %q, got %q", test.expectedErr, err.Error())
					return
				}
				return
			}
			if err == nil && test.expectedErr != "" {
				t.Errorf("expected error to be %q, got nil", test.expectedErr)
				return
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
		"equal": {
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
		"diff": {
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
		"equal-complex": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"list":   ListType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"list": List{ElemType: StringType, Elems: []attr.Value{
						String{Value: "a"},
						String{Value: "b"},
						String{Value: "c"},
					}},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"list":   ListType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"list": List{ElemType: StringType, Elems: []attr.Value{
						String{Value: "a"},
						String{Value: "b"},
						String{Value: "c"},
					}},
				},
			},
			expected: true,
		},
		"diff-complex": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"list":   ListType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"list": List{ElemType: StringType, Elems: []attr.Value{
						String{Value: "a"},
						String{Value: "b"},
						String{Value: "c"},
					}},
				},
			},
			arg: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"list":   ListType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"list": List{ElemType: StringType, Elems: []attr.Value{
						String{Value: "a"},
						String{Value: "b"},
						String{Value: "c"},
						String{Value: "d"},
					}},
				},
			},
			expected: false,
		},
		"both-unknown": {
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
		"unknown": {
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
		"both-null": {
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
		"null": {
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
		"wrong-type": {
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
		"wrong-type-complex": {
			receiver: Object{
				AttrTypes: map[string]attr.Type{
					"string": StringType,
					"list":   ListType{ElemType: StringType},
				},
				Attrs: map[string]attr.Value{
					"string": String{Value: "hello"},
					"list": List{ElemType: StringType, Elems: []attr.Value{
						String{Value: "a"},
						String{Value: "b"},
						String{Value: "c"},
					}},
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
		"diff-attribute-types": {
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
		"diff-attribute-types-count": {
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
		"diff-attribute-types-value": {
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
					"string": NumberType,
				},
				Attrs: map[string]attr.Value{
					"string": Number{Value: big.NewFloat(123)},
				},
			},
			expected: false,
		},
		"diff-attribute-count": {
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
				Attrs: map[string]attr.Value{},
			},
			expected: false,
		},
		"diff-attribute-names": {
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
