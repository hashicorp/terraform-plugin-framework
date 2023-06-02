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

func TestListNestedAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     metaschema.ListNestedAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to ListNestedAttribute"),
		},
		"ElementKeyInt": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step: tftypes.ElementKeyInt(1),
			expected: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			expectedError: nil,
		},
		"ElementKeyString": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to ListNestedAttribute"),
		},
		"ElementKeyValue": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to ListNestedAttribute"),
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

func TestListNestedAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: metaschema.ListNestedAttribute{
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

func TestListNestedAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other:    testschema.AttributeWithListValidators{},
			expected: false,
		},
		"different-attributes": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.BoolAttribute{},
					},
				},
			},
			expected: false,
		},
		"equal": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other: metaschema.ListNestedAttribute{
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

func TestListNestedAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  string
	}{
		"no-description": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"description": {
			attribute: metaschema.ListNestedAttribute{
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

func TestListNestedAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"markdown-description": {
			attribute: metaschema.ListNestedAttribute{
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

func TestListNestedAttributeGetNestedObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  metaschema.NestedAttributeObject
	}{
		"nested-object": {
			attribute: metaschema.ListNestedAttribute{
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

func TestListNestedAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"testattr": types.StringType,
					},
				},
			},
		},
		// "custom-type": {
		// 	attribute: metaschema.ListNestedAttribute{
		// 		CustomType: testtypes.ListType{},
		// 	},
		// 	expected: testtypes.ListType{},
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

func TestListNestedAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: metaschema.ListNestedAttribute{
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

func TestListNestedAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"optional": {
			attribute: metaschema.ListNestedAttribute{
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

func TestListNestedAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  bool
	}{
		"not-required": {
			attribute: metaschema.ListNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"required": {
			attribute: metaschema.ListNestedAttribute{
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

func TestListNestedAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.ListNestedAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: metaschema.ListNestedAttribute{
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
