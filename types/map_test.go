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

func TestMapTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    MapType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"map-of-strings": {
			input: MapType{
				ElemType: StringType,
			},
			expected: tftypes.Map{
				ElementType: tftypes.String,
			},
		},
		"map-of-map-of-strings": {
			input: MapType{
				ElemType: MapType{
					ElemType: StringType,
				},
			},
			expected: tftypes.Map{
				ElementType: tftypes.Map{
					ElementType: tftypes.String,
				},
			},
		},
		"map-of-map-of-map-of-strings": {
			input: MapType{
				ElemType: MapType{
					ElemType: MapType{
						ElemType: StringType,
					},
				},
			},
			expected: tftypes.Map{
				ElementType: tftypes.Map{
					ElementType: tftypes.Map{
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

func TestMapTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    MapType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"basic-map": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, map[string]tftypes.Value{
				"one":   tftypes.NewValue(tftypes.Number, 1),
				"two":   tftypes.NewValue(tftypes.Number, 2),
				"three": tftypes.NewValue(tftypes.Number, 3),
			}),
			expected: Map{
				ElemType: NumberType,
				Elems: map[string]attr.Value{
					"one":   Number{Value: big.NewFloat(1)},
					"two":   Number{Value: big.NewFloat(2)},
					"three": Number{Value: big.NewFloat(3)},
				},
			},
		},
		"wrong-type": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input:       tftypes.NewValue(tftypes.String, "wrong"),
			expectedErr: `can't use tftypes.String<"wrong"> as value of Map, can only use tftypes.Map values`,
		},
		"nil-type": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input: tftypes.NewValue(nil, nil),
			expected: Map{
				ElemType: NumberType,
				Null:     true,
			},
		},
		"unknown": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, tftypes.UnknownValue),
			expected: Map{
				ElemType: NumberType,
				Unknown:  true,
			},
		},
		"null": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, nil),
			expected: Map{
				ElemType: NumberType,
				Null:     true,
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

func TestMapTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver MapType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: MapType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
			input: MapType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
			expected: true,
		},
		"diff": {
			receiver: MapType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
			input: MapType{
				ElemType: ListType{
					ElemType: NumberType,
				},
			},
			expected: false,
		},
		"wrongType": {
			receiver: MapType{
				ElemType: StringType,
			},
			input:    NumberType,
			expected: false,
		},
		"nil": {
			receiver: MapType{
				ElemType: StringType,
			},
			input:    nil,
			expected: false,
		},
		"nil-elem": {
			receiver: MapType{},
			input:    MapType{},
			// MapTypes with nil ElemTypes are invalid, and aren't
			// equal to anything
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

func TestMapElementsAs_mapStringString(t *testing.T) {
	t.Parallel()

	var stringSlice map[string]string
	expected := map[string]string{
		"h": "hello",
		"w": "world",
	}

	diags := (Map{
		ElemType: StringType,
		Elems: map[string]attr.Value{
			"h": String{Value: "hello"},
			"w": String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestMapElementsAs_mapStringAttributeValue(t *testing.T) {
	t.Parallel()

	var stringSlice map[string]String
	expected := map[string]String{
		"h": {Value: "hello"},
		"w": {Value: "world"},
	}

	diags := (Map{
		ElemType: StringType,
		Elems: map[string]attr.Value{
			"h": String{Value: "hello"},
			"w": String{Value: "world"},
		}}).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestMapToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Map
		expectation tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"h": tftypes.NewValue(tftypes.String, "hello"),
				"w": tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"unknown": {
			input:       Map{ElemType: StringType, Unknown: true},
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"null": {
			input:       Map{ElemType: StringType, Null: true},
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		},
		"partial-unknown": {
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"unk": String{Unknown: true},
					"hw":  String{Value: "hello, world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"unk": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"hw":  tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"partial-null": {
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"n":  String{Null: true},
					"hw": String{Value: "hello, world"},
				},
			},
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"n":  tftypes.NewValue(tftypes.String, nil),
				"hw": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"no-elem-type": {
			input: Map{
				Elems: map[string]attr.Value{
					"n":  String{Null: true},
					"hw": String{Value: "hello, world"},
				},
			},
			expectedErr: "cannot convert Map to tftypes.Value if ElemType field is not set",
			expectation: tftypes.Value{},
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

func TestMapEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver Map
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			expected: true,
		},
		"elem-value-diff": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "goodnight"},
					"w": String{Value: "moon"},
				},
			},
			expected: false,
		},
		"elem-key-diff": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"no": String{Value: "hello"},
					"w":  String{Value: "world"},
				},
			},
			expected: false,
		},
		"elem-count-diff": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
					"t": String{Value: "test"},
				},
			},
			expected: false,
		},
		"elem-value-type-diff": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input: Map{
				ElemType: BoolType,
				Elems: map[string]attr.Value{
					"h": Bool{Value: false},
					"w": Bool{Value: true},
				},
			},
			expected: false,
		},
		"map-value-unknown": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input:    Map{Unknown: true},
			expected: false,
		},
		"map-value-null": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input:    Map{Null: true},
			expected: false,
		},
		"map-elem-wrongType": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"value-nil": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			input:    nil,
			expected: false,
		},
		"partially-known": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Unknown: true},
				},
			},
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Unknown: true},
				},
			},
			expected: true,
		},
		"partially-known-value-diff": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Unknown: true},
				},
			},
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			expected: false,
		},
		"partially-known-map-value-unknown": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Unknown: true},
				},
			},
			input:    Map{Unknown: true},
			expected: false,
		},
		"partially-known-map-value-null": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Unknown: true},
				},
			},
			input:    Map{Null: true},
			expected: false,
		},
		"partially-known-map-value-wrongType": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Unknown: true},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"partially-known-map-value-nil": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Unknown: true},
				},
			},
			input:    nil,
			expected: false,
		},
		"partially-null-map-value-map-value": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Null: true},
				},
			},
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Null: true},
				},
			},
			expected: true,
		},
		"partially-null-map-value-diff": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Null: true},
				},
			},
			input: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Value: "world"},
				},
			},
			expected: false,
		},
		"partially-null-map-value-unknown": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Null: true},
				},
			},
			input: Map{
				Unknown: true,
			},
			expected: false,
		},
		"partially-null-map-value-null": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Null: true},
				},
			},
			input: Map{
				Null: true,
			},
			expected: false,
		},
		"partially-null-map-value-wrongType": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Null: true},
				},
			},
			input:    String{Value: "hello, world"},
			expected: false,
		},
		"partially-null-map-value-nil": {
			receiver: Map{
				ElemType: StringType,
				Elems: map[string]attr.Value{
					"h": String{Value: "hello"},
					"w": String{Null: true},
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

func TestMapString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Map
		expectation string
	}
	tests := map[string]testCase{
		"simple": {
			input: Map{
				ElemType: Int64Type,
				Elems: map[string]attr.Value{
					"alpha": Int64{Value: 1234},
					"beta":  Int64{Value: 56789},
					"gamma": Int64{Value: 9817},
					"sigma": Int64{Value: 62534},
				},
			},
			expectation: `{"alpha":1234,"beta":56789,"gamma":9817,"sigma":62534}`,
		},
		"map-of-maps": {
			input: Map{
				ElemType: MapType{
					ElemType: StringType,
				},
				Elems: map[string]attr.Value{
					"first": Map{
						ElemType: StringType,
						Elems: map[string]attr.Value{
							"alpha": String{Value: "hello"},
							"beta":  String{Value: "world"},
							"gamma": String{Value: "foo"},
							"sigma": String{Value: "bar"},
						},
					},
					"second": Map{
						ElemType: Int64Type,
						Elems: map[string]attr.Value{
							"x": Int64{Value: 0},
							"y": Int64{Value: 0},
							"z": Int64{Value: 0},
							"t": Int64{Value: 0},
						},
					},
				},
			},
			expectation: `{"first":{"alpha":"hello","beta":"world","gamma":"foo","sigma":"bar"},"second":{"t":0,"x":0,"y":0,"z":0}}`,
		},
		"key-quotes": {
			input: Map{
				ElemType: BoolType,
				Elems: map[string]attr.Value{
					`testing is "fun"`: Bool{Value: true},
				},
			},
			expectation: `{"testing is \"fun\"":true}`,
		},
		"unknown": {
			input:       Map{Unknown: true},
			expectation: "<unknown>",
		},
		"null": {
			input:       Map{Null: true},
			expectation: "<null>",
		},
		"default-empty": {
			input:       Map{},
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
				ElemType: StringType,
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
				ElemType: StringType,
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
