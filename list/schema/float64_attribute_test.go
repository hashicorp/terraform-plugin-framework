// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFloat64AttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.Float64Attribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     schema.Float64Attribute{},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.AttributeName to basetypes.Float64Type"),
		},
		"ElementKeyInt": {
			attribute:     schema.Float64Attribute{},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.Float64Type"),
		},
		"ElementKeyString": {
			attribute:     schema.Float64Attribute{},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.Float64Type"),
		},
		"ElementKeyValue": {
			attribute:     schema.Float64Attribute{},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.Float64Type"),
		},
	}

	for name, testCase := range testCases {
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

func TestFloat64AttributeFloat64Validators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  []validator.Float64
	}{
		"no-validators": {
			attribute: schema.Float64Attribute{},
			expected:  nil,
		},
		"validators": {
			attribute: schema.Float64Attribute{
				Validators: []validator.Float64{},
			},
			expected: []validator.Float64{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Float64Validators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.Float64Attribute{},
			expected:  "",
		},
		"deprecation-message": {
			attribute: schema.Float64Attribute{
				DeprecationMessage: "test deprecation message",
			},
			expected: "test deprecation message",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetDeprecationMessage()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.Float64Attribute{},
			other:     testschema.AttributeWithFloat64Validators{},
			expected:  false,
		},
		"equal": {
			attribute: schema.Float64Attribute{},
			other:     schema.Float64Attribute{},
			expected:  true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Equal(testCase.other)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  string
	}{
		"no-description": {
			attribute: schema.Float64Attribute{},
			expected:  "",
		},
		"description": {
			attribute: schema.Float64Attribute{
				Description: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.Float64Attribute{},
			expected:  "",
		},
		"markdown-description": {
			attribute: schema.Float64Attribute{
				MarkdownDescription: "test description",
			},
			expected: "test description",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetMarkdownDescription()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.Float64Attribute{},
			expected:  types.Float64Type,
		},
		"custom-type": {
			attribute: schema.Float64Attribute{
				CustomType: testtypes.Float64Type{},
			},
			expected: testtypes.Float64Type{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetType()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.Float64Attribute{},
			expected:  false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsComputed()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.Float64Attribute{},
			expected:  false,
		},
		"optional": {
			attribute: schema.Float64Attribute{
				Optional: true,
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsOptional()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.Float64Attribute{},
			expected:  false,
		},
		"required": {
			attribute: schema.Float64Attribute{
				Required: true,
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsRequired()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.Float64Attribute{},
			expected:  false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsSensitive()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat54AttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: schema.Float64Attribute{},
			expected:  false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsWriteOnly()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: schema.Float64Attribute{},
			expected:  false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsRequiredForImport()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64AttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float64Attribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: schema.Float64Attribute{},
			expected:  false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.IsOptionalForImport()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
