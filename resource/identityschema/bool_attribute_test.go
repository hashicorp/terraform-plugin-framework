// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package identityschema_test

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
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBoolAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     identityschema.BoolAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     identityschema.BoolAttribute{},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.AttributeName to basetypes.BoolType"),
		},
		"ElementKeyInt": {
			attribute:     identityschema.BoolAttribute{},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.BoolType"),
		},
		"ElementKeyString": {
			attribute:     identityschema.BoolAttribute{},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.BoolType"),
		},
		"ElementKeyValue": {
			attribute:     identityschema.BoolAttribute{},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.BoolType"),
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

func TestBoolAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: identityschema.BoolAttribute{},
			expected:  "",
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

func TestBoolAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: identityschema.BoolAttribute{},
			other:     testschema.AttributeWithBoolValidators{},
			expected:  false,
		},
		"equal": {
			attribute: identityschema.BoolAttribute{},
			other:     identityschema.BoolAttribute{},
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

func TestBoolAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  string
	}{
		"no-description": {
			attribute: identityschema.BoolAttribute{},
			expected:  "",
		},
		"description": {
			attribute: identityschema.BoolAttribute{
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

func TestBoolAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: identityschema.BoolAttribute{},
			expected:  "",
		},
		"markdown-description-from-description": {
			attribute: identityschema.BoolAttribute{
				Description: "test description",
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

func TestBoolAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: identityschema.BoolAttribute{},
			expected:  types.BoolType,
		},
		"custom-type": {
			attribute: identityschema.BoolAttribute{
				CustomType: testtypes.BoolType{},
			},
			expected: testtypes.BoolType{},
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

func TestBoolAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: identityschema.BoolAttribute{},
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

func TestBoolAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: identityschema.BoolAttribute{},
			expected:  false,
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

func TestBoolAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  bool
	}{
		"not-required": {
			attribute: identityschema.BoolAttribute{},
			expected:  false,
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

func TestBoolAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: identityschema.BoolAttribute{},
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

func TestBoolAttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: identityschema.BoolAttribute{},
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

func TestBoolAttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: identityschema.BoolAttribute{},
			expected:  false,
		},
		"requiredForImport": {
			attribute: identityschema.BoolAttribute{
				RequiredForImport: true,
			},
			expected: true,
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

func TestBoolAttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.BoolAttribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: identityschema.BoolAttribute{},
			expected:  false,
		},
		"optionalForImport": {
			attribute: identityschema.BoolAttribute{
				OptionalForImport: true,
			},
			expected: true,
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
