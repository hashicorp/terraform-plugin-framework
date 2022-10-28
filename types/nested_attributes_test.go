package types_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListNestedAttributesType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		listNestedAttributes types.ListNestedAttributes
		expected             attr.Type
	}{
		"tfsdk-attribute": {
			listNestedAttributes: types.ListNestedAttributes{
				UnderlyingAttributes: map[string]types.Attribute{
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
		mapNestedAttributes types.MapNestedAttributes
		expected            attr.Type
	}{
		"tfsdk-attribute": {
			mapNestedAttributes: types.MapNestedAttributes{
				UnderlyingAttributes: map[string]types.Attribute{
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
		setNestedAttributes types.SetNestedAttributes
		expected            attr.Type
	}{
		"tfsdk-attribute": {
			setNestedAttributes: types.SetNestedAttributes{
				UnderlyingAttributes: map[string]types.Attribute{
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
		singleNestedAttributes types.SingleNestedAttributes
		expected               attr.Type
	}{
		"tfsdk-attribute": {
			singleNestedAttributes: types.SingleNestedAttributes{
				UnderlyingAttributes: map[string]types.Attribute{
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
