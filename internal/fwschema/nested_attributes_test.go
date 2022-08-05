package fwschema_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListNestedAttributesType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		listNestedAttributes fwschema.ListNestedAttributes
		expected             attr.Type
	}{
		"tfsdk-attribute": {
			listNestedAttributes: fwschema.ListNestedAttributes{
				UnderlyingAttributes: map[string]fwschema.Attribute{
					"test_nested_attribute": tfsdk.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_nested_attribute": types.StringType,
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.listNestedAttributes.Type()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMapNestedAttributesType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		mapNestedAttributes fwschema.MapNestedAttributes
		expected            attr.Type
	}{
		"tfsdk-attribute": {
			mapNestedAttributes: fwschema.MapNestedAttributes{
				UnderlyingAttributes: map[string]fwschema.Attribute{
					"test_nested_attribute": tfsdk.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			expected: types.MapType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_nested_attribute": types.StringType,
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.mapNestedAttributes.Type()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributesType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		setNestedAttributes fwschema.SetNestedAttributes
		expected            attr.Type
	}{
		"tfsdk-attribute": {
			setNestedAttributes: fwschema.SetNestedAttributes{
				UnderlyingAttributes: map[string]fwschema.Attribute{
					"test_nested_attribute": tfsdk.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			expected: types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"test_nested_attribute": types.StringType,
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.setNestedAttributes.Type()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedAttributesType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		singleNestedAttributes fwschema.SingleNestedAttributes
		expected               attr.Type
	}{
		"tfsdk-attribute": {
			singleNestedAttributes: fwschema.SingleNestedAttributes{
				UnderlyingAttributes: map[string]fwschema.Attribute{
					"test_nested_attribute": tfsdk.Attribute{
						Required: true,
						Type:     types.StringType,
					},
				},
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"test_nested_attribute": types.StringType,
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.singleNestedAttributes.Type()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
