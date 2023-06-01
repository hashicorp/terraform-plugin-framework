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

func TestSetTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    SetType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"set-of-strings": {
			input: SetType{
				ElemType: StringType{},
			},
			expected: tftypes.Set{
				ElementType: tftypes.String,
			},
		},
		"set-of-set-of-strings": {
			input: SetType{
				ElemType: SetType{
					ElemType: StringType{},
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
						ElemType: StringType{},
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
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			expected: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("world"),
				},
			),
		},
		"set-of-duplicate-strings": {
			receiver: SetType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "hello"),
			}),
			// Duplicate validation does not occur during this method.
			// This is okay, as tftypes allows duplicates.
			expected: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringValue("hello"),
				},
			),
		},
		"unknown-set": {
			receiver: SetType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, tftypes.UnknownValue),
			expected: NewSetUnknown(StringType{}),
		},
		"partially-unknown-set": {
			receiver: SetType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
			expected: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringUnknown(),
				},
			),
		},
		"null-set": {
			receiver: SetType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, nil),
			expected: NewSetNull(StringType{}),
		},
		"partially-null-set": {
			receiver: SetType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, nil),
			}),
			expected: NewSetValueMust(
				StringType{},
				[]attr.Value{
					NewStringValue("hello"),
					NewStringNull(),
				},
			),
		},
		"wrong-type": {
			receiver: SetType{
				ElemType: StringType{},
			},
			input:       tftypes.NewValue(tftypes.String, "wrong"),
			expectedErr: `can't use tftypes.String<"wrong"> as value of Set with ElementType basetypes.StringType, can only use tftypes.String values`,
		},
		"wrong-element-type": {
			receiver: SetType{
				ElemType: StringType{},
			},
			input: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.Number,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.Number, 1),
			}),
			expectedErr: `can't use tftypes.Set[tftypes.Number]<tftypes.Number<"1">> as value of Set with ElementType basetypes.StringType, can only use tftypes.String values`,
		},
		"nil-type": {
			receiver: SetType{
				ElemType: StringType{},
			},
			input:    tftypes.NewValue(nil, nil),
			expected: NewSetNull(StringType{}),
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
			receiver: SetType{ElemType: StringType{}},
			input:    SetType{ElemType: StringType{}},
			expected: true,
		},
		"diff": {
			receiver: SetType{ElemType: StringType{}},
			input:    SetType{ElemType: NumberType{}},
			expected: false,
		},
		"wrongType": {
			receiver: SetType{ElemType: StringType{}},
			input:    NumberType{},
			expected: false,
		},
		"nil": {
			receiver: SetType{ElemType: StringType{}},
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
