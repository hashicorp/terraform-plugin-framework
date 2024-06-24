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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFloat32AttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.Float32Attribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     schema.Float32Attribute{},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.AttributeName to basetypes.Float32Type"),
		},
		"ElementKeyInt": {
			attribute:     schema.Float32Attribute{},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.Float32Type"),
		},
		"ElementKeyString": {
			attribute:     schema.Float32Attribute{},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.Float32Type"),
		},
		"ElementKeyValue": {
			attribute:     schema.Float32Attribute{},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.Float32Type"),
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

func TestFloat32AttributeFloat32DefaultValue(t *testing.T) {
	t.Parallel()

	opt := cmp.Comparer(func(x, y defaults.Float32) bool {
		ctx := context.Background()
		req := defaults.Float32Request{}

		xResp := defaults.Float32Response{}
		x.DefaultFloat32(ctx, req, &xResp)

		yResp := defaults.Float32Response{}
		y.DefaultFloat32(ctx, req, &yResp)

		return xResp.PlanValue.Equal(yResp.PlanValue)
	})

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  defaults.Float32
	}{
		"no-default": {
			attribute: schema.Float32Attribute{},
			expected:  nil,
		},
		"default": {
			attribute: schema.Float32Attribute{
				Default: float32default.StaticFloat32(1.2345),
			},
			expected: float32default.StaticFloat32(1.2345),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Float32DefaultValue()

			if diff := cmp.Diff(got, testCase.expected, opt); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat32AttributeFloat32PlanModifiers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  []planmodifier.Float32
	}{
		"no-planmodifiers": {
			attribute: schema.Float32Attribute{},
			expected:  nil,
		},
		"planmodifiers": {
			attribute: schema.Float32Attribute{
				PlanModifiers: []planmodifier.Float32{},
			},
			expected: []planmodifier.Float32{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Float32PlanModifiers()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat32AttributeFloat32Validators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  []validator.Float32
	}{
		"no-validators": {
			attribute: schema.Float32Attribute{},
			expected:  nil,
		},
		"validators": {
			attribute: schema.Float32Attribute{
				Validators: []validator.Float32{},
			},
			expected: []validator.Float32{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Float32Validators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat32AttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.Float32Attribute{},
			expected:  "",
		},
		"deprecation-message": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.Float32Attribute{},
			other:     testschema.AttributeWithFloat32Validators{},
			expected:  false,
		},
		"equal": {
			attribute: schema.Float32Attribute{},
			other:     schema.Float32Attribute{},
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

func TestFloat32AttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  string
	}{
		"no-description": {
			attribute: schema.Float32Attribute{},
			expected:  "",
		},
		"description": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.Float32Attribute{},
			expected:  "",
		},
		"markdown-description": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.Float32Attribute{},
			expected:  types.Float32Type,
		},
		// "custom-type": {
		// 	attribute: schema.Float32Attribute{
		// 		CustomType: testtypes.Float32Type{},
		// 	},
		// 	expected: testtypes.Float32Type{},
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

func TestFloat32AttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.Float32Attribute{},
			expected:  false,
		},
		"computed": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.Float32Attribute{},
			expected:  false,
		},
		"optional": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.Float32Attribute{},
			expected:  false,
		},
		"required": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.Float32Attribute{},
			expected:  false,
		},
		"sensitive": {
			attribute: schema.Float32Attribute{
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

func TestFloat32AttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Float32Attribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"computed": {
			attribute: schema.Float32Attribute{
				Computed: true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"default-without-computed": {
			attribute: schema.Float32Attribute{
				Default: float32default.StaticFloat32(1.2),
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
			attribute: schema.Float32Attribute{
				Computed: true,
				Default:  float32default.StaticFloat32(1.2),
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
