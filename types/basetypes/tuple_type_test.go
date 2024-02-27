// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestTupleTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver TupleType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
					ObjectType{
						AttrTypes: map[string]attr.Type{
							"testattr": NumberType{},
						},
					},
				},
			},
			input: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
					ObjectType{
						AttrTypes: map[string]attr.Type{
							"testattr": NumberType{},
						},
					},
				},
			},
			expected: true,
		},
		"diff-order": {
			receiver: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
					NumberType{},
				},
			},
			input: TupleType{
				ElemTypes: []attr.Type{
					NumberType{},
					StringType{},
				},
			},
			expected: false,
		},
		"diff-length": {
			receiver: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
					StringType{},
				},
			},
			input: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
				},
			},
			expected: false,
		},
		"wrong-type": {
			receiver: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
				},
			},
			input:    ListType{ElemType: StringType{}},
			expected: false,
		},
		"nil": {
			receiver: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
				},
			},
			input:    nil,
			expected: false,
		},
		"nil-elems": {
			receiver: TupleType{},
			input:    TupleType{},
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

func TestTupleTypeString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    TupleType
		expected string
	}{
		"ElemTypes-empty": {
			input: TupleType{
				ElemTypes: []attr.Type{},
			},
			expected: "types.TupleType[]",
		},
		"ElemTypes-one": {
			input: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
				},
			},
			expected: "types.TupleType[basetypes.StringType]",
		},
		"ElemTypes-multiple": {
			input: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
					ObjectType{
						AttrTypes: map[string]attr.Type{
							"testattr": NumberType{},
						},
					},
				},
			},
			expected: "types.TupleType[basetypes.StringType, types.ObjectType[\"testattr\":basetypes.NumberType]]",
		},
		"ElemTypes-missing": {
			input:    TupleType{},
			expected: "types.TupleType[]", // intentionally similar to empty
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestTupleTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    TupleType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"tuple-of-strings": {
			input: TupleType{
				ElemTypes: []attr.Type{
					StringType{},
				},
			},
			expected: tftypes.Tuple{
				ElementTypes: []tftypes.Type{
					tftypes.String,
				},
			},
		},
		"tuple-of-tuple-of-strings": {
			input: TupleType{
				ElemTypes: []attr.Type{
					TupleType{
						ElemTypes: []attr.Type{
							StringType{},
						},
					},
				},
			},
			expected: tftypes.Tuple{
				ElementTypes: []tftypes.Type{
					tftypes.Tuple{
						ElementTypes: []tftypes.Type{
							tftypes.String,
						},
					},
				},
			},
		},
		"tuple-of-tuple-of-object": {
			input: TupleType{
				ElemTypes: []attr.Type{
					TupleType{
						ElemTypes: []attr.Type{
							ObjectType{
								AttrTypes: map[string]attr.Type{
									"testattr": NumberType{},
								},
							},
						},
					},
				},
			},
			expected: tftypes.Tuple{
				ElementTypes: []tftypes.Type{
					tftypes.Tuple{
						ElementTypes: []tftypes.Type{
							tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"testattr": tftypes.Number,
								},
							},
						},
					},
				},
			},
		},
		"ElemTypes-empty": {
			input: TupleType{
				ElemTypes: make([]attr.Type, 0),
			},
			expected: tftypes.Tuple{
				ElementTypes: make([]tftypes.Type, 0),
			},
		},
		"ElemTypes-missing": {
			input: TupleType{},
			expected: tftypes.Tuple{
				ElementTypes: make([]tftypes.Type, 0),
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

func TestTupleTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    TupleType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"tuple-with-same-types": {
			receiver: TupleType{
				[]attr.Type{StringType{}, StringType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			expected: NewTupleValueMust(
				[]attr.Type{StringType{}, StringType{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
		},
		"tuple-with-multiple-types": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}, BoolType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.Bool, true),
			}),
			expected: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewBoolValue(true),
				},
			),
		},
		"tuple-with-dynamic-types": {
			receiver: TupleType{
				[]attr.Type{DynamicType{}, DynamicType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.DynamicPseudoType, tftypes.DynamicPseudoType},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.Bool, true),
			}),
			expected: NewTupleValueMust(
				[]attr.Type{DynamicType{}, DynamicType{}},
				[]attr.Value{
					NewDynamicValue(NewStringValue("hello")),
					NewDynamicValue(NewBoolValue(true)),
				},
			),
		},
		"unknown-tuple": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}, BoolType{}, DynamicType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType},
			}, tftypes.UnknownValue),
			expected: NewTupleUnknown([]attr.Type{StringType{}, BoolType{}, DynamicType{}}),
		},
		"partially-unknown-tuple": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}, BoolType{}, DynamicType{}, DynamicType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType, tftypes.DynamicPseudoType},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
				tftypes.NewValue(tftypes.String, "world"),
				tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue),
			}),
			expected: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}, DynamicType{}, DynamicType{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewBoolUnknown(),
					NewDynamicValue(NewStringValue("world")),
					NewDynamicUnknown(),
				},
			),
		},
		"null-tuple": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}, BoolType{}, DynamicType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType},
			}, nil),
			expected: NewTupleNull([]attr.Type{StringType{}, BoolType{}, DynamicType{}}),
		},
		"partially-null-tuple": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}, BoolType{}, DynamicType{}, DynamicType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.Bool, tftypes.DynamicPseudoType, tftypes.DynamicPseudoType},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.Bool, nil),
				tftypes.NewValue(tftypes.String, "world"),
				tftypes.NewValue(tftypes.DynamicPseudoType, nil),
			}),
			expected: NewTupleValueMust(
				[]attr.Type{StringType{}, BoolType{}, DynamicType{}, DynamicType{}},
				[]attr.Value{
					NewStringValue("hello"),
					NewBoolNull(),
					NewDynamicValue(NewStringValue("world")),
					NewDynamicNull(),
				},
			),
		},
		"wrong-type": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}},
			},
			input:       tftypes.NewValue(tftypes.Bool, true),
			expectedErr: "expected tftypes.Tuple[tftypes.String], got tftypes.Bool",
		},
		"wrong-element-types": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}, BoolType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.Bool, tftypes.String},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.String, "hello"),
			}),
			expectedErr: "expected tftypes.Tuple[tftypes.String, tftypes.Bool], got tftypes.Tuple[tftypes.Bool, tftypes.String]",
		},
		"mismatched-element-types-length": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}, StringType{}},
			},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String, tftypes.String},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
				tftypes.NewValue(tftypes.String, "invalid"),
			}),
			expectedErr: "expected tftypes.Tuple[tftypes.String, tftypes.String], got tftypes.Tuple[tftypes.String, tftypes.String, tftypes.String]",
		},
		"nil-type": {
			receiver: TupleType{
				ElemTypes: []attr.Type{StringType{}},
			},
			input:    tftypes.NewValue(nil, nil),
			expected: NewTupleNull([]attr.Type{StringType{}}),
		},
		"missing-element-type": {
			receiver: TupleType{},
			input: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: make([]tftypes.Type, 0),
			}, make([]tftypes.Value, 0)),
			expected: NewTupleValueMust(
				make([]attr.Type, 0),
				make([]attr.Value, 0),
			),
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
