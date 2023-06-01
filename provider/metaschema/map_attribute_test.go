// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package metaschema_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestMapAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     metaschema.MapAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     metaschema.MapAttribute{ElementType: types.StringType},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to MapType"),
		},
		"ElementKeyInt": {
			attribute:     metaschema.MapAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to MapType"),
		},
		"ElementKeyString": {
			attribute:     metaschema.MapAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyString("test"),
			expected:      types.StringType,
			expectedError: nil,
		},
		"ElementKeyValue": {
			attribute:     metaschema.MapAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to MapType"),
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

func TestMapAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			expected:  "",
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

func TestMapAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			other:     testschema.AttributeWithMapValidators{},
			expected:  false,
		},
		"different-element-type": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			other:     metaschema.MapAttribute{ElementType: types.BoolType},
			expected:  false,
		},
		"equal": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			other:     metaschema.MapAttribute{ElementType: types.StringType},
			expected:  true,
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

func TestMapAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		expected  string
	}{
		"no-description": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"description": {
			attribute: metaschema.MapAttribute{
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

func TestMapAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"markdown-description": {
			attribute: metaschema.MapAttribute{
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

func TestMapAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			expected:  types.MapType{ElemType: types.StringType},
		},
		// "custom-type": {
		// 	attribute: metaschema.MapAttribute{
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

func TestMapAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			expected:  false,
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

func TestMapAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"optional": {
			attribute: metaschema.MapAttribute{
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

func TestMapAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		expected  bool
	}{
		"not-required": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"required": {
			attribute: metaschema.MapAttribute{
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

func TestMapAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: metaschema.MapAttribute{ElementType: types.StringType},
			expected:  false,
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

func TestMapAttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute metaschema.MapAttribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"customtype": {
			attribute: metaschema.MapAttribute{
				CustomType: testtypes.MapType{},
				Optional:   true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype": {
			attribute: metaschema.MapAttribute{
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
			attribute: metaschema.MapAttribute{
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
		name, testCase := name, testCase

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
