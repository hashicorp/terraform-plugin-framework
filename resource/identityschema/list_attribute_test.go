// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identityschema_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     identityschema.ListAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     identityschema.ListAttribute{ElementType: types.StringType},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to ListType"),
		},
		"ElementKeyInt": {
			attribute:     identityschema.ListAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyInt(1),
			expected:      types.StringType,
			expectedError: nil,
		},
		"ElementKeyString": {
			attribute:     identityschema.ListAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to ListType"),
		},
		"ElementKeyValue": {
			attribute:     identityschema.ListAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to ListType"),
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

func TestListAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
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

func TestListAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
			other:     testschema.AttributeWithListValidators{},
			expected:  false,
		},
		"different-element-type": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
			other:     identityschema.ListAttribute{ElementType: types.BoolType},
			expected:  false,
		},
		"equal": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
			other:     identityschema.ListAttribute{ElementType: types.StringType},
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

func TestListAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  string
	}{
		"no-description": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"description": {
			attribute: identityschema.ListAttribute{
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

func TestListAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"markdown-description-from-description": {
			attribute: identityschema.ListAttribute{
				ElementType: types.StringType,
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

func TestListAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
			expected:  types.ListType{ElemType: types.StringType},
		},
		"custom-type": {
			attribute: identityschema.ListAttribute{
				CustomType: testtypes.ListType{ListType: types.ListType{ElemType: types.StringType}},
			},
			expected: testtypes.ListType{ListType: types.ListType{ElemType: types.StringType}},
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

func TestListAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
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

func TestListAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
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

func TestListAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  bool
	}{
		"not-required": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
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

func TestListAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
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

func TestListAttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
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

func TestListAttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"requiredForImport": {
			attribute: identityschema.ListAttribute{
				ElementType:       types.StringType,
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

func TestListAttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: identityschema.ListAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"optionalForImport": {
			attribute: identityschema.ListAttribute{
				ElementType:       types.StringType,
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

func TestListAttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute identityschema.ListAttribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"elementtype": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.StringType,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-missing": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
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
		"elementtype-bool": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.BoolType,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-int64": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.Int64Type,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-int32": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.Int32Type,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-float64": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.Float64Type,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-float32": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.Float32Type,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-number": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.NumberType,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-notprimitive-dynamic": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.DynamicType,
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
							"\"test\" contains an Attribute of type \"basetypes.DynamicType\" that is not allowed for Lists in Resource Identity. "+
							"Lists in Resource Identity may only have primitive element types such as Bool, Int, Float, Number and String.",
					),
				},
			},
		},
		"elementtype-notprimitive-object": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.ObjectType{},
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
							"\"test\" contains an element of type \"types.ObjectType[]\" that is not allowed for Lists in Resource Identity. "+
							"Lists in Resource Identity may only have primitive element types such as Bool, Int, Float, Number and String.",
					),
				},
			},
		},
		"elementtype-notprimitive-map": {
			attribute: identityschema.ListAttribute{
				RequiredForImport: true,
				ElementType:       types.MapType{},
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
							"\"test\" contains an Attribute of type \"types.MapType[!!! MISSING TYPE !!!]\" that is not allowed for Lists in Resource Identity. "+
							"Lists in Resource Identity may only have primitive element types such as Bool, Int, Float, Number and String.",
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
