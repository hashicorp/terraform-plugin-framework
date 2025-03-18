// Copyright (c) HashiCorp, Inc.
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

func TestInt32AttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     identityschema.Int32Attribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     identityschema.Int32Attribute{},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.AttributeName to basetypes.Int32Type"),
		},
		"ElementKeyInt": {
			attribute:     identityschema.Int32Attribute{},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.Int32Type"),
		},
		"ElementKeyString": {
			attribute:     identityschema.Int32Attribute{},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.Int32Type"),
		},
		"ElementKeyValue": {
			attribute:     identityschema.Int32Attribute{},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.Int32Type"),
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

func TestInt32AttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: identityschema.Int32Attribute{},
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

func TestInt32AttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: identityschema.Int32Attribute{},
			other:     testschema.AttributeWithInt32Validators{},
			expected:  false,
		},
		"equal": {
			attribute: identityschema.Int32Attribute{},
			other:     identityschema.Int32Attribute{},
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

func TestInt32AttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  string
	}{
		"no-description": {
			attribute: identityschema.Int32Attribute{},
			expected:  "",
		},
		"description": {
			attribute: identityschema.Int32Attribute{
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

func TestInt32AttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: identityschema.Int32Attribute{},
			expected:  "",
		},
		"markdown-description-from-description": {
			attribute: identityschema.Int32Attribute{
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

func TestInt32AttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  attr.Type
	}{
		"base": {
			attribute: identityschema.Int32Attribute{},
			expected:  types.Int32Type,
		},
		"custom-type": {
			attribute: identityschema.Int32Attribute{
				CustomType: testtypes.Int32Type{},
			},
			expected: testtypes.Int32Type{},
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

func TestInt32AttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  bool
	}{
		"not-computed": {
			attribute: identityschema.Int32Attribute{},
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

func TestInt32AttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  bool
	}{
		"not-optional": {
			attribute: identityschema.Int32Attribute{},
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

func TestInt32AttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  bool
	}{
		"not-required": {
			attribute: identityschema.Int32Attribute{},
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

func TestInt32AttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: identityschema.Int32Attribute{},
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

func TestInt32AttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: identityschema.Int32Attribute{},
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

func TestInt32AttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: identityschema.Int32Attribute{},
			expected:  false,
		},
		"requiredForImport": {
			attribute: identityschema.Int32Attribute{
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

func TestInt32AttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int32Attribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: identityschema.Int32Attribute{},
			expected:  false,
		},
		"optionalForImport": {
			attribute: identityschema.Int32Attribute{
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
