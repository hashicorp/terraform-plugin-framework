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

func TestInt64AttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     identityschema.Int64Attribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     identityschema.Int64Attribute{},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.AttributeName to basetypes.Int64Type"),
		},
		"ElementKeyInt": {
			attribute:     identityschema.Int64Attribute{},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.Int64Type"),
		},
		"ElementKeyString": {
			attribute:     identityschema.Int64Attribute{},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.Int64Type"),
		},
		"ElementKeyValue": {
			attribute:     identityschema.Int64Attribute{},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.Int64Type"),
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

func TestInt64AttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: identityschema.Int64Attribute{},
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

func TestInt64AttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: identityschema.Int64Attribute{},
			other:     testschema.AttributeWithInt64Validators{},
			expected:  false,
		},
		"equal": {
			attribute: identityschema.Int64Attribute{},
			other:     identityschema.Int64Attribute{},
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

func TestInt64AttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  string
	}{
		"no-description": {
			attribute: identityschema.Int64Attribute{},
			expected:  "",
		},
		"description": {
			attribute: identityschema.Int64Attribute{
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

func TestInt64AttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: identityschema.Int64Attribute{},
			expected:  "",
		},
		"markdown-description-from-description": {
			attribute: identityschema.Int64Attribute{
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

func TestInt64AttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  attr.Type
	}{
		"base": {
			attribute: identityschema.Int64Attribute{},
			expected:  types.Int64Type,
		},
		"custom-type": {
			attribute: identityschema.Int64Attribute{
				CustomType: testtypes.Int64Type{},
			},
			expected: testtypes.Int64Type{},
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

func TestInt64AttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  bool
	}{
		"not-computed": {
			attribute: identityschema.Int64Attribute{},
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

func TestInt64AttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  bool
	}{
		"not-optional": {
			attribute: identityschema.Int64Attribute{},
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

func TestInt64AttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  bool
	}{
		"not-required": {
			attribute: identityschema.Int64Attribute{},
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

func TestInt64AttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: identityschema.Int64Attribute{},
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

func TestInt64AttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: identityschema.Int64Attribute{},
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

func TestInt64AttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: identityschema.Int64Attribute{},
			expected:  false,
		},
		"requiredForImport": {
			attribute: identityschema.Int64Attribute{
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

func TestInt64AttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.Int64Attribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: identityschema.Int64Attribute{},
			expected:  false,
		},
		"optionalForImport": {
			attribute: identityschema.Int64Attribute{
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
