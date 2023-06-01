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

func TestMapNestedAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     metaschema.MapNestedAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to MapNestedAttribute"),
		},
		"ElementKeyInt": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to MapNestedAttribute"),
		},
		"ElementKeyString": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step: tftypes.ElementKeyString("test"),
			expected: metaschema.NestedAttributeObject{
				Attributes: map[string]metaschema.Attribute{
					"testattr": metaschema.StringAttribute{},
				},
			},
			expectedError: nil,
		},
		"ElementKeyValue": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to MapNestedAttribute"),
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

func TestMapNestedAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: metaschema.MapNestedAttribute{
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

func TestMapNestedAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other:    testschema.AttributeWithMapValidators{},
			expected: false,
		},
		"different-attributes": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.BoolAttribute{},
					},
				},
			},
			expected: false,
		},
		"equal": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			other: metaschema.MapNestedAttribute{
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

func TestMapNestedAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  string
	}{
		"no-description": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"description": {
			attribute: metaschema.MapNestedAttribute{
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

func TestMapNestedAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"markdown-description": {
			attribute: metaschema.MapNestedAttribute{
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

func TestMapNestedAttributeGetNestedObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  metaschema.NestedAttributeObject
	}{
		"nested-object": {
			attribute: metaschema.MapNestedAttribute{
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

func TestMapNestedAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: types.MapType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"testattr": types.StringType,
					},
				},
			},
		},
		// "custom-type": {
		// 	attribute: metaschema.MapNestedAttribute{
		// 		CustomType: testtypes.MapType{},
		// 	},
		// 	expected: testtypes.MapType{},
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

func TestMapNestedAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: metaschema.MapNestedAttribute{
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

func TestMapNestedAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"optional": {
			attribute: metaschema.MapNestedAttribute{
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

func TestMapNestedAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  bool
	}{
		"not-required": {
			attribute: metaschema.MapNestedAttribute{
				NestedObject: metaschema.NestedAttributeObject{
					Attributes: map[string]metaschema.Attribute{
						"testattr": metaschema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"required": {
			attribute: metaschema.MapNestedAttribute{
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

func TestMapNestedAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapNestedAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: metaschema.MapNestedAttribute{
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
