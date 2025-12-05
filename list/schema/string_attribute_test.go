// Copyright IBM Corp. 2021, 2025
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
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     listschema.StringAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     listschema.StringAttribute{},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.AttributeName to basetypes.StringType"),
		},
		"ElementKeyInt": {
			attribute:     listschema.StringAttribute{},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.StringType"),
		},
		"ElementKeyString": {
			attribute:     listschema.StringAttribute{},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.StringType"),
		},
		"ElementKeyValue": {
			attribute:     listschema.StringAttribute{},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.StringType"),
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

func TestStringAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: listschema.StringAttribute{},
			expected:  "",
		},
		"deprecation-message": {
			attribute: listschema.StringAttribute{
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

func TestStringAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: listschema.StringAttribute{},
			other:     testschema.AttributeWithStringValidators{},
			expected:  false,
		},
		"equal": {
			attribute: listschema.StringAttribute{},
			other:     listschema.StringAttribute{},
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

func TestStringAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  string
	}{
		"no-description": {
			attribute: listschema.StringAttribute{},
			expected:  "",
		},
		"description": {
			attribute: listschema.StringAttribute{
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

func TestStringAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: listschema.StringAttribute{},
			expected:  "",
		},
		"markdown-description": {
			attribute: listschema.StringAttribute{
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

func TestStringAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: listschema.StringAttribute{},
			expected:  types.StringType,
		},
		"custom-type": {
			attribute: listschema.StringAttribute{
				CustomType: testtypes.StringType{},
			},
			expected: testtypes.StringType{},
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

func TestStringAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: listschema.StringAttribute{},
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

func TestStringAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: listschema.StringAttribute{},
			expected:  false,
		},
		"optional": {
			attribute: listschema.StringAttribute{
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

func TestStringAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  bool
	}{
		"not-required": {
			attribute: listschema.StringAttribute{},
			expected:  false,
		},
		"required": {
			attribute: listschema.StringAttribute{
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

func TestStringAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: listschema.StringAttribute{},
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

func TestStringAttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: listschema.StringAttribute{},
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

func TestStringAttributeStringValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  []validator.String
	}{
		"no-validators": {
			attribute: listschema.StringAttribute{},
			expected:  nil,
		},
		"validators": {
			attribute: listschema.StringAttribute{
				Validators: []validator.String{},
			},
			expected: []validator.String{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.StringValidators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestStringAttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: listschema.StringAttribute{},
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

func TestStringAttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute listschema.StringAttribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: listschema.StringAttribute{},
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
