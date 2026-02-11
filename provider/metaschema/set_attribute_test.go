// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package metaschema_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     metaschema.SetAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     metaschema.SetAttribute{ElementType: types.StringType},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to SetType"),
		},
		"ElementKeyInt": {
			attribute:     metaschema.SetAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to SetType"),
		},
		"ElementKeyString": {
			attribute:     metaschema.SetAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to SetType"),
		},
		"ElementKeyValue": {
			attribute:     metaschema.SetAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      types.StringType,
			expectedError: nil,
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

func TestSetAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
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

func TestSetAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
			other:     testschema.AttributeWithSetValidators{},
			expected:  false,
		},
		"different-element-type": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
			other:     metaschema.SetAttribute{ElementType: types.BoolType},
			expected:  false,
		},
		"equal": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
			other:     metaschema.SetAttribute{ElementType: types.StringType},
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

func TestSetAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  string
	}{
		"no-description": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"description": {
			attribute: metaschema.SetAttribute{
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

func TestSetAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"markdown-description": {
			attribute: metaschema.SetAttribute{
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

func TestSetAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
			expected:  types.SetType{ElemType: types.StringType},
		},
		"custom-type": {
			attribute: metaschema.SetAttribute{
				CustomType: testtypes.SetType{SetType: types.SetType{ElemType: types.StringType}},
			},
			expected: testtypes.SetType{SetType: types.SetType{ElemType: types.StringType}},
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

func TestSetAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
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

func TestSetAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"optional": {
			attribute: metaschema.SetAttribute{
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

func TestSetAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  bool
	}{
		"not-required": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"required": {
			attribute: metaschema.SetAttribute{
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

func TestSetAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: metaschema.SetAttribute{ElementType: types.StringType},
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

func TestSetAttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: metaschema.SetAttribute{},
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

func TestSetAttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"customtype": {
			attribute: metaschema.SetAttribute{
				CustomType: testtypes.SetType{},
				Optional:   true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype": {
			attribute: metaschema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-missing": {
			attribute: metaschema.SetAttribute{
				Optional: true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Attribute Implementation",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"test\" is missing the CustomType or ElementType field on a collection Attribute. "+
							"One of these fields is required to prevent other unexpected errors or panics.",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwschema.ValidateImplementationResponse{}
			testCase.attribute.ValidateImplementation(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSetAttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: metaschema.SetAttribute{},
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

func TestSetAttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.SetAttribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: metaschema.SetAttribute{},
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
