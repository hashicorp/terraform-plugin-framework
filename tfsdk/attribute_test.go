package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute Attribute
		expected  attr.Type
	}{
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
		"Type-NumberType": {
			attribute: Attribute{
				Required: true,
				Type:     types.NumberType,
			},
			expected: types.NumberType,
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

func TestAttributeTerraformType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute Attribute
		expected  tftypes.Type
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
			expected: tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_nested_attribute": tftypes.String,
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
			expected: tftypes.Map{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_nested_attribute": tftypes.String,
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
			expected: tftypes.Set{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_nested_attribute": tftypes.String,
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
			expected: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test_nested_attribute": tftypes.String,
				},
			},
		},
		"Type-BoolType": {
			attribute: Attribute{
				Required: true,
				Type:     types.BoolType,
			},
			expected: tftypes.Bool,
		},
		"Type-Float64Type": {
			attribute: Attribute{
				Required: true,
				Type:     types.Float64Type,
			},
			expected: tftypes.Number,
		},
		"Type-Int64Type": {
			attribute: Attribute{
				Required: true,
				Type:     types.Int64Type,
			},
			expected: tftypes.Number,
		},
		"Type-ListType": {
			attribute: Attribute{
				Required: true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
			},
			expected: tftypes.List{
				ElementType: tftypes.String,
			},
		},
		"Type-MapType": {
			attribute: Attribute{
				Required: true,
				Type: types.MapType{
					ElemType: types.StringType,
				},
			},
			expected: tftypes.Map{
				ElementType: tftypes.String,
			},
		},
		"Type-NumberType": {
			attribute: Attribute{
				Required: true,
				Type:     types.NumberType,
			},
			expected: tftypes.Number,
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
			expected: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test_object_attribute": tftypes.String,
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
			expected: tftypes.Set{
				ElementType: tftypes.String,
			},
		},
		"Type-StringType": {
			attribute: Attribute{
				Required: true,
				Type:     types.StringType,
			},
			expected: tftypes.String,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.terraformType(context.Background())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
