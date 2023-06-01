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
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSingleNestedBlockApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block         schema.SingleNestedBlock
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName-attribute": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("testattr"),
			expected:      schema.StringAttribute{},
			expectedError: nil,
		},
		"AttributeName-block": {
			block: schema.SingleNestedBlock{
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
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("other"),
			expected:      nil,
			expectedError: fmt.Errorf("no attribute or block \"other\" on SingleNestedBlock"),
		},
		"ElementKeyInt": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to SingleNestedBlock"),
		},
		"ElementKeyString": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to SingleNestedBlock"),
		},
		"ElementKeyValue": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to SingleNestedBlock"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.block.ApplyTerraform5AttributePathStep(testCase.step)

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

func TestSingleNestedBlockGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    schema.SingleNestedBlock
		expected string
	}{
		"no-deprecation-message": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"deprecation-message": {
			block: schema.SingleNestedBlock{
				DeprecationMessage: "test deprecation message",
			},
			expected: "test deprecation message",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.GetDeprecationMessage()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedBlockEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    schema.SingleNestedBlock
		other    fwschema.Block
		expected bool
	}{
		"different-type": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			other:    testschema.BlockWithObjectValidators{},
			expected: false,
		},
		"different-attributes": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			other: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.BoolAttribute{},
				},
			},
			expected: false,
		},
		"equal": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			other: schema.SingleNestedBlock{
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

			got := testCase.block.Equal(testCase.other)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedBlockGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    schema.SingleNestedBlock
		expected string
	}{
		"no-description": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"description": {
			block: schema.SingleNestedBlock{
				Description: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.GetDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedBlockGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    schema.SingleNestedBlock
		expected string
	}{
		"no-markdown-description": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"markdown-description": {
			block: schema.SingleNestedBlock{
				MarkdownDescription: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.GetMarkdownDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedBlockGetNestedObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    schema.SingleNestedBlock
		expected schema.NestedBlockObject
	}{
		"nested-object": {
			block: schema.SingleNestedBlock{
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
			expected: schema.NestedBlockObject{
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
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.GetNestedObject()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedBlockObjectPlanModifiers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    schema.SingleNestedBlock
		expected []planmodifier.Object
	}{
		"no-planmodifiers": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: nil,
		},
		"planmodifiers": {
			block: schema.SingleNestedBlock{
				PlanModifiers: []planmodifier.Object{},
			},
			expected: []planmodifier.Object{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.ObjectPlanModifiers()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedBlockObjectValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    schema.SingleNestedBlock
		expected []validator.Object
	}{
		"no-validators": {
			block: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: nil,
		},
		"validators": {
			block: schema.SingleNestedBlock{
				Validators: []validator.Object{},
			},
			expected: []validator.Object{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.ObjectValidators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedBlockType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		block    schema.SingleNestedBlock
		expected attr.Type
	}{
		"base": {
			block: schema.SingleNestedBlock{
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
		// 	block: schema.SingleNestedBlock{
		// 		CustomType: testtypes.SingleType{},
		// 	},
		// 	expected: testtypes.SingleType{},
		// },
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.block.Type()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
