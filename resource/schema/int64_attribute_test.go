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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestInt64AttributeApplyTerraform5AttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute     schema.Int64Attribute
		step          tftypes.AttributePathStep
		expected      any
		expectedError error
	}{
		"AttributeName": {
			attribute:     schema.Int64Attribute{},
			step:          tftypes.AttributeName("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.AttributeName to basetypes.Int64Type"),
		},
		"ElementKeyInt": {
			attribute:     schema.Int64Attribute{},
			step:          tftypes.ElementKeyInt(1),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyInt to basetypes.Int64Type"),
		},
		"ElementKeyString": {
			attribute:     schema.Int64Attribute{},
			step:          tftypes.ElementKeyString("test"),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyString to basetypes.Int64Type"),
		},
		"ElementKeyValue": {
			attribute:     schema.Int64Attribute{},
			step:          tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			expected:      nil,
			expectedError: fmt.Errorf("cannot apply AttributePathStep tftypes.ElementKeyValue to basetypes.Int64Type"),
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

func TestInt64AttributeGetDeprecationMessage(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  string
	}{
		"no-deprecation-message": {
			attribute: schema.Int64Attribute{},
			expected:  "",
		},
		"deprecation-message": {
			attribute: schema.Int64Attribute{
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

func TestInt64AttributeEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		other     fwschema.Attribute
		expected  bool
	}{
		"different-type": {
			attribute: schema.Int64Attribute{},
			other:     testschema.AttributeWithInt64Validators{},
			expected:  false,
		},
		"equal": {
			attribute: schema.Int64Attribute{},
			other:     schema.Int64Attribute{},
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

func TestInt64AttributeGetDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  string
	}{
		"no-description": {
			attribute: schema.Int64Attribute{},
			expected:  "",
		},
		"description": {
			attribute: schema.Int64Attribute{
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

func TestInt64AttributeGetMarkdownDescription(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  string
	}{
		"no-markdown-description": {
			attribute: schema.Int64Attribute{},
			expected:  "",
		},
		"markdown-description": {
			attribute: schema.Int64Attribute{
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

func TestInt64AttributeGetType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  attr.Type
	}{
		"base": {
			attribute: schema.Int64Attribute{},
			expected:  types.Int64Type,
		},
		// "custom-type": {
		// 	attribute: schema.Int64Attribute{
		// 		CustomType: testtypes.Int64Type{},
		// 	},
		// 	expected: testtypes.Int64Type{},
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

func TestInt64AttributeInt64DefaultValue(t *testing.T) {
	t.Parallel()

	opt := cmp.Comparer(func(x, y defaults.Int64) bool {
		ctx := context.Background()
		req := defaults.Int64Request{}

		xResp := defaults.Int64Response{}
		x.DefaultInt64(ctx, req, &xResp)

		yResp := defaults.Int64Response{}
		y.DefaultInt64(ctx, req, &yResp)

		return xResp.PlanValue.Equal(yResp.PlanValue)
	})

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  defaults.Int64
	}{
		"no-default": {
			attribute: schema.Int64Attribute{},
			expected:  nil,
		},
		"default": {
			attribute: schema.Int64Attribute{
				Default: int64default.StaticInt64(12345),
			},
			expected: int64default.StaticInt64(12345),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Int64DefaultValue()

			if diff := cmp.Diff(got, testCase.expected, opt); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestInt64AttributeInt64PlanModifiers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  []planmodifier.Int64
	}{
		"no-planmodifiers": {
			attribute: schema.Int64Attribute{},
			expected:  nil,
		},
		"planmodifiers": {
			attribute: schema.Int64Attribute{
				PlanModifiers: []planmodifier.Int64{},
			},
			expected: []planmodifier.Int64{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Int64PlanModifiers()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestInt64AttributeInt64Validators(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  []validator.Int64
	}{
		"no-validators": {
			attribute: schema.Int64Attribute{},
			expected:  nil,
		},
		"validators": {
			attribute: schema.Int64Attribute{
				Validators: []validator.Int64{},
			},
			expected: []validator.Int64{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.attribute.Int64Validators()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestInt64AttributeIsComputed(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  bool
	}{
		"not-computed": {
			attribute: schema.Int64Attribute{},
			expected:  false,
		},
		"computed": {
			attribute: schema.Int64Attribute{
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

func TestInt64AttributeIsOptional(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  bool
	}{
		"not-optional": {
			attribute: schema.Int64Attribute{},
			expected:  false,
		},
		"optional": {
			attribute: schema.Int64Attribute{
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

func TestInt64AttributeIsRequired(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  bool
	}{
		"not-required": {
			attribute: schema.Int64Attribute{},
			expected:  false,
		},
		"required": {
			attribute: schema.Int64Attribute{
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

func TestInt64AttributeIsSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		expected  bool
	}{
		"not-sensitive": {
			attribute: schema.Int64Attribute{},
			expected:  false,
		},
		"sensitive": {
			attribute: schema.Int64Attribute{
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

func TestInt64AttributeValidateImplementation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attribute schema.Int64Attribute
		request   fwschema.ValidateImplementationRequest
		expected  *fwschema.ValidateImplementationResponse
	}{
		"computed": {
			attribute: schema.Int64Attribute{
				Computed: true,
			},
			request: fwschema.ValidateImplementationRequest{
				Name: "test",
				Path: path.Root("test"),
			},
			expected: &fwschema.ValidateImplementationResponse{},
		},
		"default-without-computed": {
			attribute: schema.Int64Attribute{
				Default: int64default.StaticInt64(123),
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
			attribute: schema.Int64Attribute{
				Computed: true,
				Default:  int64default.StaticInt64(123),
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
