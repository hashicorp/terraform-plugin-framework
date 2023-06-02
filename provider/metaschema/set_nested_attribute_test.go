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
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSetNestedAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     metaschema.SetNestedAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to SetNestedAttribute"),
		},
		"ElementKeyInt": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to SetNestedAttribute"),
		},
		"ElementKeyString": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to SetNestedAttribute"),
		},
		"ElementKeyValue": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step: tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
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
		attribute metaschema.SetNestedAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: "",
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
		attribute metaschema.SetNestedAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other:    testschema.AttributeWithSetValidators{},
			expected: false,
		},
		"different-attributes": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.BoolAttribute{},
					},
				},
			},
			expected: false,
		},
		"equal": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
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
		attribute metaschema.SetNestedAttribute
		expected  string
	}{
		"no-description": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"description": {
			attribute: metaschema.SetNestedAttribute{
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
		attribute metaschema.SetNestedAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"markdown-description": {
			attribute: metaschema.SetNestedAttribute{
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
		attribute metaschema.SetNestedAttribute
		expected  metaschema.NestedAttributeObject
	}{
		"nested-object": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
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
		attribute metaschema.SetNestedAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
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
		// 	attribute: metaschema.SetNestedAttribute{
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
		attribute metaschema.SetNestedAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
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
		attribute metaschema.SetNestedAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"optional": {
			attribute: metaschema.SetNestedAttribute{
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
		attribute metaschema.SetNestedAttribute
		expected  bool
	}{
		"not-required": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"required": {
			attribute: metaschema.SetNestedAttribute{
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
		attribute metaschema.SetNestedAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: metaschema.SetNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
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

			got := testCase.attribute.IsSensitive()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
