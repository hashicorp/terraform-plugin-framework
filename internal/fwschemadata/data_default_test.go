package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataDefault(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data          *fwschemadata.Data
		expected      *fwschemadata.Data
		expectedDiags diag.Diagnostics
	}{
		"bool-attribute-unmodified": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, true),
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, true),
					},
				),
			},
		},
		"bool-attribute-null": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // intentionally nil
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // intentionally nil
					},
				),
			},
		},
		"bool-attribute-null-default": {
			data: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Optional: true,
							Default:  booldefault.StaticValue(false),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, nil), // intentionally nil
					},
				),
			},
			expected: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"bool_attribute": testschema.AttributeWithBoolDefaultValue{
							Optional: true,
							Default:  booldefault.StaticValue(false),
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"bool_attribute": tftypes.Bool,
						},
					},
					map[string]tftypes.Value{
						"bool_attribute": tftypes.NewValue(tftypes.Bool, false), // intentionally nil
					},
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.data.TransformPlanDefaults(context.Background())

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(testCase.data, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
