// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func TestTypeCanBeStringKeyed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType attr.Type
		expected    bool
	}{
		"nil": {
			elementType: nil,
			expected:    false,
		},
		"primitive": {
			elementType: StringType{},
			expected:    true,
		},
		"object": {
			elementType: ObjectType{
				AttrTypes: map[string]attr.Type{
					"bool":   BoolType{},
					"string": StringType{},
				},
			},
			expected: true,
		},
		"list-of-object": {
			elementType: ListType{
				ElemType: ObjectType{
					AttrTypes: map[string]attr.Type{
						"string": StringType{},
					},
				},
			},
			expected: true,
		},
		"map-of-list": {
			elementType: MapType{
				ElemType: ListType{ElemType: StringType{}},
			},
			expected: true,
		},
		"tuple": {
			elementType: TupleType{
				ElemTypes: []attr.Type{StringType{}, BoolType{}},
			},
			expected: true,
		},
		"set": {
			elementType: SetType{ElemType: StringType{}},
			expected:    false,
		},
		"dynamic": {
			elementType: DynamicType{},
			expected:    false,
		},
		"object-with-nested-set": {
			elementType: ObjectType{
				AttrTypes: map[string]attr.Type{
					"string": StringType{},
					"set":    SetType{ElemType: StringType{}},
				},
			},
			expected: false,
		},
		"list-of-object-with-nested-set": {
			elementType: ListType{
				ElemType: ObjectType{
					AttrTypes: map[string]attr.Type{
						"set": SetType{ElemType: StringType{}},
					},
				},
			},
			expected: false,
		},
		"tuple-with-nested-dynamic": {
			elementType: TupleType{
				ElemTypes: []attr.Type{StringType{}, DynamicType{}},
			},
			expected: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := typeCanBeStringKeyed(testCase.elementType)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}
