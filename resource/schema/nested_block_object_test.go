// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestNestedBlockObjectApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object        schema.NestedBlockObject
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName-attribute": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("testattr"),
			expected:      schema.StringAttribute{},
			expectedError: nil,
		},
		"AttributeName-block": {
			object: schema.NestedBlockObject{
				Blocks: map[string]schema.Block{
					"testblock": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
				},
			},
			step: tftypes.AttributeName("testblock"),
			expected: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expectedError: nil,
		},
		"AttributeName-missing": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("other"),
			expected:      nil,
			expectedError: fmt.Errorf("no attribute or block \"other\" on NestedBlockObject"),
		},
		"ElementKeyInt": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to NestedBlockObject"),
		},
		"ElementKeyString": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to NestedBlockObject"),
		},
		"ElementKeyValue": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to NestedBlockObject"),
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

func TestNestedBlockObjectEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object   schema.NestedBlockObject
		other    fwschema.NestedBlockObject
		expected bool
	}{
		"different-attributes": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			other: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.BoolAttribute{},
				},
			},
			expected: false,
		},
		"equal": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			other: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
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

func TestNestedBlockObjectGetAttributes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object   schema.NestedBlockObject
		expected fwschema.UnderlyingAttributes
	}{
		"no-attributes": {
			object:   schema.NestedBlockObject{},
			expected: fwschema.UnderlyingAttributes{},
		},
		"attributes": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr1": schema.StringAttribute{},
					"testattr2": schema.StringAttribute{},
				},
			},
			expected: fwschema.UnderlyingAttributes{
				"testattr1": schema.StringAttribute{},
				"testattr2": schema.StringAttribute{},
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

func TestNestedBlockObjectGetBlocks(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object   schema.NestedBlockObject
		expected map[string]fwschema.Block
	}{
		"no-blocks": {
			object:   schema.NestedBlockObject{},
			expected: map[string]fwschema.Block{},
		},
		"blocks": {
			object: schema.NestedBlockObject{
				Blocks: map[string]schema.Block{
					"testblock1": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
					"testblock2": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
				},
			},
			expected: map[string]fwschema.Block{
				"testblock1": schema.SingleNestedBlock{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
				"testblock2": schema.SingleNestedBlock{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.object.GetBlocks()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedBlockObjectObjectPlanModifiers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.NestedBlockObject
		expected  []planmodifier.Object
	}{
		"no-planmodifiers": {
			attribute: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: nil,
		},
		"planmodifiers": {
			attribute: schema.NestedBlockObject{
				PlanModifiers: []planmodifier.Object{},
			},
			expected: []planmodifier.Object{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ObjectPlanModifiers()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedBlockObjectObjectValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.NestedBlockObject
		expected  []validator.Object
	}{
		"no-validators": {
			attribute: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: nil,
		},
		"validators": {
			attribute: schema.NestedBlockObject{
				Validators: []validator.Object{},
			},
			expected: []validator.Object{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ObjectValidators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNestedBlockObjectType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		object   schema.NestedBlockObject
		expected attr.Type
	}{
		"base": {
			object: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
				Blocks: map[string]schema.Block{
					"testblock": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"testattr": schema.StringAttribute{},
						},
					},
				},
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"testattr": types.StringType,
					"testblock": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"testattr": types.StringType,
						},
					},
				},
			},
		},
		// "custom-type": {
		// 	block: schema.NestedBlockObject{
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
