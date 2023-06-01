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

func TestSingleNestedAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     metaschema.SingleNestedAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("testattr"),
			expected:      metaschema.StringAttribute{},
			expectedError: nil,
		},
		"AttributeName-missing": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("other"),
			expected:      nil,
			expectedError: fmt.Errorf("no attribute \"other\" on SingleNestedAttribute"),
		},
		"ElementKeyInt": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to SingleNestedAttribute"),
		},
		"ElementKeyString": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to SingleNestedAttribute"),
		},
		"ElementKeyValue": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to SingleNestedAttribute"),
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

func TestSingleNestedAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			other:    testschema.AttributeWithObjectValidators{},
			expected: false,
		},
		"different-attributes": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			other: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.BoolAttribute{},
				},
			},
			expected: false,
		},
		"equal": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			other: metaschema.SingleNestedAttribute{
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

			got := testCase.attribute.Equal(testCase.other)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
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

func TestSingleNestedAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  string
	}{
		"no-description": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			expected: "",
		},
		"description": {
			attribute: metaschema.SingleNestedAttribute{
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

func TestSingleNestedAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			expected: "",
		},
		"markdown-description": {
			attribute: metaschema.SingleNestedAttribute{
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

func TestSingleNestedAttributeGetNestedObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  metaschema.NestedAttributeObject
	}{
		"nested-object": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
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

func TestSingleNestedAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: metaschema.SingleNestedAttribute{
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
		// 	attribute: metaschema.SingleNestedAttribute{
		// 		CustomType: testtypes.SingleType{},
		// 	},
		// 	expected: testtypes.SingleType{},
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

func TestSingleNestedAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
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

func TestSingleNestedAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			expected: false,
		},
		"optional": {
			attribute: metaschema.SingleNestedAttribute{
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

func TestSingleNestedAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  bool
	}{
		"not-required": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			expected: false,
		},
		"required": {
			attribute: metaschema.SingleNestedAttribute{
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

func TestSingleNestedAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SingleNestedAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: metaschema.SingleNestedAttribute{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
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
