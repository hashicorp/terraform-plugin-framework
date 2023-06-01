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
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSetNestedAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.SetNestedAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to SetNestedAttribute"),
		},
		"ElementKeyInt": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to SetNestedAttribute"),
		},
		"ElementKeyString": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to SetNestedAttribute"),
		},
		"ElementKeyValue": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			step: tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expectedError: nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.attribute.ApplyTerraform5AttributePathStep(testCase.step)

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

func TestSetNestedAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"deprecation-message": {
			attribute: schema.SetNestedAttribute{
				DeprecationMessage: "test deprecation message",
			},
			expected: "test deprecation message",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetDeprecationMessage()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			other:    testschema.AttributeWithSetValidators{},
			expected: false,
		},
		"different-attributes": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			other: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.BoolAttribute{},
					},
				},
			},
			expected: false,
		},
		"equal": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			other: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Equal(testCase.other)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  string
	}{
		"no-description": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"description": {
			attribute: schema.SetNestedAttribute{
				Description: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"markdown-description": {
			attribute: schema.SetNestedAttribute{
				MarkdownDescription: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetMarkdownDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeGetNestedObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  schema.NestedAttributeObject
	}{
		"nested-object": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetNestedObject()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"testattr": types.StringType,
					},
				},
			},
		},
		// "custom-type": {
		// 	attribute: schema.SetNestedAttribute{
		// 		CustomType: testtypes.SetType{},
		// 	},
		// 	expected: testtypes.SetType{},
		// },
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

func TestSetNestedAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsComputed()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"optional": {
			attribute: schema.SetNestedAttribute{
				Optional: true,
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsOptional()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"required": {
			attribute: schema.SetNestedAttribute{
				Required: true,
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsRequired()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"sensitive": {
			attribute: schema.SetNestedAttribute{
				Sensitive: true,
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsSensitive()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetNestedAttributeSetValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SetNestedAttribute
		expected  []validator.Set
	}{
		"no-validators": {
			attribute: schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: nil,
		},
		"validators": {
			attribute: schema.SetNestedAttribute{
				Validators: []validator.Set{},
			},
			expected: []validator.Set{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.SetValidators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
