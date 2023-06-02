// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package metaschema_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestNestedAttributeObjectApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object        metaschema.NestedAttributeObject
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("testattr"),
			expected:      metaschema.StringAttribute{},
			expectedError: nil,
		},
		"AttributeName-missing": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("other"),
			expected:      nil,
			expectedError: fmt.Errorf("no attribute \"other\" on NestedAttributeObject"),
		},
		"ElementKeyInt": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to NestedAttributeObject"),
		},
		"ElementKeyString": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to NestedAttributeObject"),
		},
		"ElementKeyValue": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to NestedAttributeObject"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.object.ApplyTerraform5AttributePathStep(testCase.step)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedAttributeObjectEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object   metaschema.NestedAttributeObject
		other    fwschema.NestedAttributeObject
		expected bool
	}{
		"different-attributes": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			other: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.BoolAttribute{},
				},
			},
			expected: false,
		},
		"equal": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			other: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.object.Equal(testCase.other)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedAttributeObjectGetAttributes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object   metaschema.NestedAttributeObject
		expected fwschema.UnderlyingAttributes
	}{
		"no-attributes": {
			object:   metaschema.NestedAttributeObject{},
			expected: fwschema.UnderlyingAttributes{},
		},
		"attributes": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr1": metaschema.StringAttribute{},
					"testattr2": metaschema.StringAttribute{},
				},
			},
			expected: fwschema.UnderlyingAttributes{
				"testattr1": metaschema.StringAttribute{},
				"testattr2": metaschema.StringAttribute{},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.object.GetAttributes()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedAttributeObjectType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object   metaschema.NestedAttributeObject
		expected attr.Type
	}{
		"base": {
			object: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"testattr": types.StringType,
				},
			},
		},
		// "custom-type": {
		// 	block: metaschema.NestedAttributeObject{
		// 		CustomType: testtypes.SingleType{},
		// 	},
		// 	expected: testtypes.SingleType{},
		// },
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.object.Type()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
