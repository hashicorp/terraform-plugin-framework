// Copyright IBM Corp. 2021, 2026
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

func TestFloat32AttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.Float32Attribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     schema.Float32Attribute{},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.AttributeName to basetypes.Float32Type"),
		},
		"ElementKeyInt": {
			attribute:     schema.Float32Attribute{},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.Float32Type"),
		},
		"ElementKeyString": {
			attribute:     schema.Float32Attribute{},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.Float32Type"),
		},
		"ElementKeyValue": {
			attribute:     schema.Float32Attribute{},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.Float32Type"),
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

func TestFloat32AttributeFloat32Validators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  []validator.Float32
	}{
		"no-validators": {
			attribute: schema.Float32Attribute{},
			expected:  nil,
		},
		"validators": {
			attribute: schema.Float32Attribute{
				Validators: []validator.Float32{},
			},
			expected: []validator.Float32{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Float32Validators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat32AttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.Float32Attribute{},
			expected:  "",
		},
		"deprecation-message": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.Float32Attribute{},
			other:     testschema.AttributeWithFloat32Validators{},
			expected:  false,
		},
		"equal": {
			attribute: schema.Float32Attribute{},
			other:     schema.Float32Attribute{},
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

func TestFloat32AttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  string
	}{
		"no-description": {
			attribute: schema.Float32Attribute{},
			expected:  "",
		},
		"description": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.Float32Attribute{},
			expected:  "",
		},
		"markdown-description": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.Float32Attribute{},
			expected:  types.Float32Type,
		},
		"custom-type": {
			attribute: schema.Float32Attribute{
				CustomType: testtypes.Float32Type{},
			},
			expected: testtypes.Float32Type{},
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

func TestFloat32AttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.Float32Attribute{},
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

func TestFloat32AttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.Float32Attribute{},
			expected:  false,
		},
		"optional": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.Float32Attribute{},
			expected:  false,
		},
		"required": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.Float32Attribute{},
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

func TestFloat32AttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: schema.Float32Attribute{},
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

func TestFloat32AttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: schema.Float32Attribute{},
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

func TestFloat32AttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: schema.Float32Attribute{},
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
