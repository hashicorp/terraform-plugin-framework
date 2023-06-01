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
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.ListAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     schema.ListAttribute{ElementType: types.StringType},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.AttributeName to ListType"),
		},
		"ElementKeyInt": {
			attribute:     schema.ListAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyInt(1),
			expected:      types.StringType,
			expectedError: nil,
		},
		"ElementKeyString": {
			attribute:     schema.ListAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to ListType"),
		},
		"ElementKeyValue": {
			attribute:     schema.ListAttribute{ElementType: types.StringType},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to ListType"),
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

func TestListAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"deprecation-message": {
			attribute: schema.ListAttribute{
				DeprecationMessage: "test deprecation message",
			},
			expected: "test deprecation message",
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

func TestListAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			other:     testschema.AttributeWithListValidators{},
			expected:  false,
		},
		"different-element-type": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			other:     schema.ListAttribute{ElementType: types.BoolType},
			expected:  false,
		},
		"equal": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			other:     schema.ListAttribute{ElementType: types.StringType},
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

func TestListAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  string
	}{
		"no-description": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"description": {
			attribute: schema.ListAttribute{
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

func TestListAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  "",
		},
		"markdown-description": {
			attribute: schema.ListAttribute{
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

func TestListAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  types.ListType{ElemType: types.StringType},
		},
		// "custom-type": {
		// 	attribute: schema.ListAttribute{
		// 		CustomType: testtypes.ListType{},
		// 	},
		// 	expected: testtypes.ListType{},
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

func TestListAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"computed": {
			attribute: schema.ListAttribute{
				Computed: true,
			},
			expected: true,
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

func TestListAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"optional": {
			attribute: schema.ListAttribute{
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

func TestListAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"required": {
			attribute: schema.ListAttribute{
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

func TestListAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  false,
		},
		"sensitive": {
			attribute: schema.ListAttribute{
				Sensitive: true,
			},
			expected: true,
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

func TestListAttributeListDefaultValue(t *testing.T) {
	t.Parallel()

	opt := cmp.Comparer(func(x, y defaults.List) bool {
		ctx := context.Background()
		req := defaults.ListRequest{}

		xResp := defaults.ListResponse{}
		x.DefaultList(ctx, req, &xResp)

		yResp := defaults.ListResponse{}
		y.DefaultList(ctx, req, &yResp)

		return xResp.PlanValue.Equal(yResp.PlanValue)
	})

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  defaults.List
	}{
		"no-default": {
			attribute: schema.ListAttribute{},
			expected:  nil,
		},
		"default": {
			attribute: schema.ListAttribute{
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("test-value¬"),
						},
					),
				),
			},
			expected: listdefault.StaticValue(
				types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("test-value¬"),
					},
				),
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ListDefaultValue()

			if diff := cmp.Diff(got, testCase.expected, opt); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestListAttributeListPlanModifiers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  []planmodifier.List
	}{
		"no-planmodifiers": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  nil,
		},
		"planmodifiers": {
			attribute: schema.ListAttribute{
				PlanModifiers: []planmodifier.List{},
			},
			expected: []planmodifier.List{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ListPlanModifiers()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestListAttributeListValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		expected  []validator.List
	}{
		"no-validators": {
			attribute: schema.ListAttribute{ElementType: types.StringType},
			expected:  nil,
		},
		"validators": {
			attribute: schema.ListAttribute{
				Validators: []validator.List{},
			},
			expected: []validator.List{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ListValidators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestListAttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ListAttribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"computed": {
			attribute: schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"customtype": {
			attribute: schema.ListAttribute{
				Computed:   true,
				CustomType: testtypes.ListType{},
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"default-without-computed": {
			attribute: schema.ListAttribute{
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("test"),
						},
					),
				),
				ElementType: types.StringType,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Schema Using Attribute Default For Non-Computed Attribute",
						"Attribute \"test\" must be computed when using default. "+
							"This is an issue with the provider and should be reported to the provider developers.",
					),
				},
			},
		},
		"default-with-computed": {
			attribute: schema.ListAttribute{
				Computed: true,
				Default: listdefault.StaticValue(
					types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("test"),
						},
					),
				),
				ElementType: types.StringType,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype": {
			attribute: schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"elementtype-missing": {
			attribute: schema.ListAttribute{
				Computed: true,
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
