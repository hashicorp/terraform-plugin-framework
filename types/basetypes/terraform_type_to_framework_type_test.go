// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestTerraformTypeToFrameworkType(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    tftypes.Type
		expected attr.Type
	}{
		"bool": {
			input:    tftypes.Bool,
			expected: BoolType{},
		},
		"number": {
			input:    tftypes.Number,
			expected: NumberType{},
		},
		"string": {
			input:    tftypes.String,
			expected: StringType{},
		},
		"dynamic": {
			input:    tftypes.DynamicPseudoType,
			expected: DynamicType{},
		},
		"list": {
			input:    tftypes.List{ElementType: tftypes.Bool},
			expected: ListType{ElemType: BoolType{}},
		},
		"set": {
			input:    tftypes.Set{ElementType: tftypes.Number},
			expected: SetType{ElemType: NumberType{}},
		},
		"map": {
			input:    tftypes.Map{ElementType: tftypes.String},
			expected: MapType{ElemType: StringType{}},
		},
		"object": {
			input: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"bool": tftypes.Bool,
					"list": tftypes.List{ElementType: tftypes.Number},
					"nested_obj": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"map":    tftypes.Map{ElementType: tftypes.DynamicPseudoType},
							"string": tftypes.String,
						},
					},
				},
			},
			expected: ObjectType{
				AttrTypes: map[string]attr.Type{
					"bool": BoolType{},
					"list": ListType{ElemType: NumberType{}},
					"nested_obj": ObjectType{
						AttrTypes: map[string]attr.Type{
							"map":    MapType{ElemType: DynamicType{}},
							"string": StringType{},
						},
					},
				},
			},
		},
		"tuple": {
			input: tftypes.Tuple{
				ElementTypes: []tftypes.Type{
					tftypes.Bool,
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"list":   tftypes.List{ElementType: tftypes.DynamicPseudoType},
							"number": tftypes.Number,
						},
					},
					tftypes.Map{ElementType: tftypes.String},
				},
			},
			expected: TupleType{
				ElemTypes: []attr.Type{
					BoolType{},
					ObjectType{
						AttrTypes: map[string]attr.Type{
							"list":   ListType{ElemType: DynamicType{}},
							"number": NumberType{},
						},
					},
					MapType{ElemType: StringType{}},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := TerraformTypeToFrameworkType(test.input)
			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("Unexpected diff (-expected, +got): %s", diff)
			}
		})
	}
}
