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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSingleNestedAttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.SingleNestedAttribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("testattr"),
			expected:      schema.StringAttribute{},
			expectedError: nil,
		},
		"AttributeName-missing": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.AttributeName("other"),
			expected:      nil,
			expectedError: fmt.Errorf("no attribute \"other\" on SingleNestedAttribute"),
		},
		"ElementKeyInt": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyInt to SingleNestedAttribute"),
		},
		"ElementKeyString": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyString to SingleNestedAttribute"),
		},
		"ElementKeyValue": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply step tftypes.ElementKeyValue to SingleNestedAttribute"),
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

func TestSingleNestedAttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			other:    testschema.AttributeWithObjectValidators{},
			expected: false,
		},
		"different-attributes": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			other: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.BoolAttribute{},
				},
			},
			expected: false,
		},
		"equal": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			other: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: true,
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

func TestSingleNestedAttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"deprecation-message": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  string
	}{
		"no-description": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"description": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: "",
		},
		"markdown-description": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeGetNestedObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  schema.NestedAttributeObject
	}{
		"nested-object": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
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
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.GetNestedObject()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSingleNestedAttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"testattr": types.StringType,
				},
			},
		},
		// "custom-type": {
		// 	attribute: schema.SingleNestedAttribute{
		// 		CustomType: testtypes.SingleType{},
		// 	},
		// 	expected: testtypes.SingleType{},
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

func TestSingleNestedAttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: false,
		},
		"computed": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: false,
		},
		"optional": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: false,
		},
		"required": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: false,
		},
		"sensitive": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeObjectDefaultValue(t *testing.T) {
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
		attribute schema.SingleNestedAttribute
		expected  defaults.Object
	}{
		"no-default": {
			attribute: schema.SingleNestedAttribute{},
			expected:  nil,
		},
		"default": {
			attribute: schema.SingleNestedAttribute{
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"one": types.StringType,
						},
						map[string]attr.Value{
							"one": types.StringValue("test-value"),
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
						"one": types.StringValue("test-value"),
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

func TestSingleNestedAttributeObjectPlanModifiers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  []planmodifier.Object
	}{
		"no-planmodifiers": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: nil,
		},
		"planmodifiers": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeObjectValidators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		expected  []validator.Object
	}{
		"no-validators": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"testattr": schema.StringAttribute{},
				},
			},
			expected: nil,
		},
		"validators": {
			attribute: schema.SingleNestedAttribute{
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

func TestSingleNestedAttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.SingleNestedAttribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"computed": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"test_attr": schema.StringAttribute{
						Computed: true,
					},
				},
				Computed: true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"default-without-computed": {
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"test_attr": schema.StringAttribute{
						Computed: true,
					},
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
			attribute: schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"test_attr": schema.StringAttribute{
						Computed: true,
					},
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
