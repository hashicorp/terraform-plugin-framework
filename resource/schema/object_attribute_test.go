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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestObjectAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.ObjectAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			step:          tftypes.AttributeName("testattr"),
			expected:      types.StringType,
			expectedError: nil,
		},
		"AttributeName-missing": {
			attribute:     schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			step:          tftypes.AttributeName("other"),
			expected:      nil,
			expectedError: fmt.Errorf("undefined attribute name other in ObjectType"),
		},
		"ElementKeyInt": {
			attribute:     schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to ObjectType"),
		},
		"ElementKeyString": {
			attribute:     schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to ObjectType"),
		},
		"ElementKeyValue": {
			attribute:     schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to ObjectType"),
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

func TestObjectAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  "",
		},
		"deprecation-message": {
			attribute: schema.ObjectAttribute{
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

func TestObjectAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			other:     testschema.AttributeWithObjectValidators{},
			expected:  false,
		},
		"different-attribute-type": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			other:     schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.BoolType}},
			expected:  false,
		},
		"equal": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			other:     schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
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

func TestObjectAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  string
	}{
		"no-description": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  "",
		},
		"description": {
			attribute: schema.ObjectAttribute{
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

func TestObjectAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  "",
		},
		"markdown-description": {
			attribute: schema.ObjectAttribute{
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

func TestObjectAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  types.ObjectType{AttrTypes: map[string]attr.Type{"testattr": types.StringType}},
		},
		// "custom-type": {
		// 	attribute: schema.ObjectAttribute{
		// 		CustomType: testtypes.ObjectType{},
		// 	},
		// 	expected: testtypes.ObjectType{},
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

func TestObjectAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  false,
		},
		"computed": {
			attribute: schema.ObjectAttribute{
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

func TestObjectAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  false,
		},
		"optional": {
			attribute: schema.ObjectAttribute{
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

func TestObjectAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  false,
		},
		"required": {
			attribute: schema.ObjectAttribute{
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

func TestObjectAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  false,
		},
		"sensitive": {
			attribute: schema.ObjectAttribute{
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

func TestObjectAttributeObjectDefaultValue(t *testing.T) {
	t.Parallel()

	opt := cmp.Comparer(func(x, y defaults.Object) bool {
		ctx := context.Background()
		req := defaults.ObjectRequest{}

		xResp := defaults.ObjectResponse{}
		x.DefaultObject(ctx, req, &xResp)

		yResp := defaults.ObjectResponse{}
		y.DefaultObject(ctx, req, &yResp)

		return xResp.PlanValue.Equal(yResp.PlanValue)
	})

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  defaults.Object
	}{
		"no-default": {
			attribute: schema.ObjectAttribute{},
			expected:  nil,
		},
		"default": {
			attribute: schema.ObjectAttribute{
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"one": types.StringType,
						},
						map[string]attr.Value{
							"one": types.StringValue("test-value¬"),
						},
					),
				),
			},
			expected: objectdefault.StaticValue(
				types.ObjectValueMust(
					map[string]attr.Type{
						"one": types.StringType,
					},
					map[string]attr.Value{
						"one": types.StringValue("test-value¬"),
					},
				),
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ObjectDefaultValue()

			if diff := cmp.Diff(got, testCase.expected, opt); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestObjectAttributeObjectPlanModifiers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  []planmodifier.Object
	}{
		"no-planmodifiers": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  nil,
		},
		"planmodifiers": {
			attribute: schema.ObjectAttribute{
				PlanModifiers: []planmodifier.Object{},
			},
			expected: []planmodifier.Object{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ObjectPlanModifiers()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestObjectAttributeObjectValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		expected  []validator.Object
	}{
		"no-validators": {
			attribute: schema.ObjectAttribute{AttributeTypes: map[string]attr.Type{"testattr": types.StringType}},
			expected:  nil,
		},
		"validators": {
			attribute: schema.ObjectAttribute{
				Validators: []validator.Object{},
			},
			expected: []validator.Object{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.ObjectValidators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestObjectAttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.ObjectAttribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"attributetypes": {
			attribute: schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
				Computed: true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"attributetypes-missing": {
			attribute: schema.ObjectAttribute{
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
							"\"test\" is missing the AttributeTypes or CustomType field on an object Attribute. "+
							"One of these fields is required to prevent other unexpected errors or panics.",
					),
				},
			},
		},
		"computed": {
			attribute: schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
				Computed: true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"customtype": {
			attribute: schema.ObjectAttribute{
				Computed:   true,
				CustomType: testtypes.ObjectType{},
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"default-without-computed": {
			attribute: schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("testvalue"),
						},
					),
				),
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
			attribute: schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"test_attr": types.StringType,
				},
				Computed: true,
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("testvalue"),
						},
					),
				),
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
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
