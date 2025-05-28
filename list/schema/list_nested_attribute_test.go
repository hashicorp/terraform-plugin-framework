// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema_test

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
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListNestedAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.ListNestedAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to ListNestedAttribute"),
		},
		"ElementKeyInt": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			step: tftypes.ElementKeyInt(1),
			expected: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expectedError: nil,
		},
		"ElementKeyString": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to ListNestedAttribute"),
		},
		"ElementKeyValue": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to ListNestedAttribute"),
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

func TestListNestedAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"deprecation-message": {
			attribute: schema.ListNestedAttribute{
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

func TestListNestedAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			other:    testschema.AttributeWithListValidators{},
			expected: false,
		},
		"different-attributes-definitions": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			other: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			expected: false,
		},
		"different-attributes-types": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			other: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.BoolAttribute{},
					},
				},
			},
			expected: false,
		},
		"equal": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			other: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: true,
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

func TestListNestedAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  string
	}{
		"no-description": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"description": {
			attribute: schema.ListNestedAttribute{
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

func TestListNestedAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: "",
		},
		"markdown-description": {
			attribute: schema.ListNestedAttribute{
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

func TestListNestedAttributeGetNestedObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  schema.NestedAttributeObject
	}{
		"nested-object": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetNestedObject()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestListNestedAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"testattr": types.StringType,
					},
				},
			},
		},
		"custom-type": {
			attribute: schema.ListNestedAttribute{
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

func TestListNestedAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: false,
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

func TestListNestedAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"optional": {
			attribute: schema.ListNestedAttribute{
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

func TestListNestedAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"required": {
			attribute: schema.ListNestedAttribute{
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

func TestListNestedAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: false,
		},
		"sensitive": {
			attribute: schema.ListNestedAttribute{
				Sensitive: true,
			},
			expected: true,
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

func TestListNestedAttributeIsWriteOnly(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  bool
	}{
		"not-writeOnly": {
			attribute: schema.ListNestedAttribute{},
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

func TestListNestedAttributeListValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  []validator.List
	}{
		"no-validators": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"testattr": schema.StringAttribute{},
					},
				},
			},
			expected: nil,
		},
		"validators": {
			attribute: schema.ListNestedAttribute{
				Validators: []validator.List{},
			},
			expected: []validator.List{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ListValidators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestListNestedAttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"customtype": {
			attribute: schema.ListNestedAttribute{
				Optional:   true,
				CustomType: testtypes.ListType{},
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"nestedobject": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"test_attr": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"nestedobject-dynamic": {
			attribute: schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"test_dyn": schema.DynamicAttribute{
							Optional: true,
						},
					},
				},
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Schema Implementation",
						"When validating the schema, an implementation issue was found. "+
							"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
							"\"test\" is an attribute that contains a collection type with a nested dynamic type.\n\n"+
							"Dynamic types inside of collections are not currently supported in terraform-plugin-framework. "+
							"If underlying dynamic values are required, replace the \"test\" attribute definition with DynamicAttribute instead.",
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

func TestListNestedAttributeIsRequiredForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  bool
	}{
		"not-requiredForImport": {
			attribute: schema.ListNestedAttribute{},
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

func TestListNestedAttributeIsOptionalForImport(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListNestedAttribute
		expected  bool
	}{
		"not-optionalForImport": {
			attribute: schema.ListNestedAttribute{},
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
