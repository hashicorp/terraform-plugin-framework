package schema_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
