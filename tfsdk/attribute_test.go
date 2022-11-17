package tfsdk

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAttributeGetNestedMode(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute Attribute
		expected  fwschema.NestingMode
	}{
		"unset": {
			attribute: Attribute{
				Type: types.StringType,
			},
			expected: fwschema.NestingModeUnknown,
		},
		"list": {
			attribute: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{}),
			},
			expected: fwschema.NestingModeList,
		},
		"map": {
			attribute: Attribute{
				Attributes: MapNestedAttributes(map[string]Attribute{}),
			},
			expected: fwschema.NestingModeMap,
		},
		"set": {
			attribute: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{}),
			},
			expected: fwschema.NestingModeSet,
		},
		"single": {
			attribute: Attribute{
				Attributes: SingleNestedAttributes(map[string]Attribute{}),
			},
			expected: fwschema.NestingModeSingle,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetNestingMode()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute Attribute
		expected  attr.Type
	}{
		"Attributes-ListNestedAttributes": {
			attribute: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"test_nested_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				}),
				Required: true,
			},
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_nested_attribute": types.StringType,
					},
				},
			},
		},
		"Attributes-MapNestedAttributes": {
			attribute: Attribute{
				Attributes: MapNestedAttributes(map[string]Attribute{
					"test_nested_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				}),
				Required: true,
			},
			expected: types.MapType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_nested_attribute": types.StringType,
					},
				},
			},
		},
		"Attributes-SetNestedAttributes": {
			attribute: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"test_nested_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				}),
				Required: true,
			},
			expected: types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_nested_attribute": types.StringType,
					},
				},
			},
		},
		"Attributes-SingleNestedAttributes": {
			attribute: Attribute{
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"test_nested_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				}),
				Required: true,
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_nested_attribute": types.StringType,
				},
			},
		},
		"Type-BoolType": {
			attribute: Attribute{
				Required: true,
				Type:     types.BoolType,
			},
			expected: types.BoolType,
		},
		"Type-Float64Type": {
			attribute: Attribute{
				Required: true,
				Type:     types.Float64Type,
			},
			expected: types.Float64Type,
		},
		"Type-Int64Type": {
			attribute: Attribute{
				Required: true,
				Type:     types.Int64Type,
			},
			expected: types.Int64Type,
		},
		"Type-ListType": {
			attribute: Attribute{
				Required: true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			expected: types.ListType{
				ElemType: types.StringType,
			},
		},
		"Type-MapType": {
			attribute: Attribute{
				Required: true,
				Type: types.MapType{
					ElemType: types.StringType,
				},
			},
			expected: types.MapType{
				ElemType: types.StringType,
			},
		},
		"Type-NumberType": {
			attribute: Attribute{
				Required: true,
				Type:     types.NumberType,
			},
			expected: types.NumberType,
		},
		"Type-ObjectType": {
			attribute: Attribute{
				Required: true,
				Type: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_object_attribute": types.StringType,
					},
				},
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_object_attribute": types.StringType,
				},
			},
		},
		"Type-SetType": {
			attribute: Attribute{
				Required: true,
				Type: types.SetType{
					ElemType: types.StringType,
				},
			},
			expected: types.SetType{
				ElemType: types.StringType,
			},
		},
		"Type-StringType": {
			attribute: Attribute{
				Required: true,
				Type:     types.StringType,
			},
			expected: types.StringType,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetType()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
