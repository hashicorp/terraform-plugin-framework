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

func TestListTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    ListType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"list-of-strings": {
			input: ListType{
				ElemType: StringType{},
			},
			expected: tftypes.List{
				ElementType: tftypes.String,
			},
		},
		"list-of-list-of-strings": {
			input: ListType{
				ElemType: ListType{
					ElemType: StringType{},
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
						ElemType: StringType{},
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
			t.Parallel()

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
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
		},
		"unknown-list": {
			receiver: ListType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: NewListUnknown(StringType{}),
		},
		"partially-unknown-list": {
			receiver: ListType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringUnknown(),
				},
			),
		},
		"null-list": {
			receiver: ListType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, nil),
			expected: NewListNull(StringType{}),
		},
		"partially-null-list": {
			receiver: ListType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, nil),
			}),
			expected: NewListValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringNull(),
				},
			),
		},
		"wrong-type": {
			receiver: ListType{
				ElemType: StringType{},
			},
			input:       tftypes.NewValue(tftypes.String, "wrong"),
			expectedErr: `can't use tftypes.String<"wrong"> as value of List with ElementType basetypes.StringType, can only use tftypes.String values`,
		},
		"wrong-element-type": {
			receiver: ListType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.Number,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 1),
			}),
			expectedErr: `can't use tftypes.List[tftypes.Number]<tftypes.Number<"1">> as value of List with ElementType basetypes.StringType, can only use tftypes.String values`,
		},
		"nil-type": {
			receiver: ListType{
				ElemType: StringType{},
			},
			input:    tftypes.NewValue(nil, nil),
			expected: NewListNull(StringType{}),
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

func TestListTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver ListType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: ListType{ElemType: StringType{}},
			input:    ListType{ElemType: StringType{}},
			expected: true,
		},
		"diff": {
			receiver: ListType{ElemType: StringType{}},
			input:    ListType{ElemType: NumberType{}},
			expected: false,
		},
		"wrongType": {
			receiver: ListType{ElemType: StringType{}},
			input:    NumberType{},
			expected: false,
		},
		"nil": {
			receiver: ListType{ElemType: StringType{}},
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
