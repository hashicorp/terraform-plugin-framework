package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestSchemaProposeNewState(t *testing.T) {
	tests := map[string]struct {
		schema      fwschema.Schema
		priorVal    map[string]tftypes.Value
		configVal   map[string]tftypes.Value
		expectedVal map[string]tftypes.Value
	}{
		"empty": {
			schema:      schema.Schema{},
			priorVal:    map[string]tftypes.Value{},
			configVal:   map[string]tftypes.Value{},
			expectedVal: map[string]tftypes.Value{},
		},
		"no prior": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"optional_attribute": schema.StringAttribute{
						Optional: true,
					},
					"computed_attribute": schema.StringAttribute{
						Computed: true,
					},
					"single_nested_attribute": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"required_nested_attribute": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				Blocks: map[string]schema.Block{
					"single_nested_block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"optional_computed_attributeA": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							"optional_computed_attributeB": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
						},
					},
				},
			},
			priorVal: nil,
			configVal: map[string]tftypes.Value{
				"optional_attribute":      tftypes.NewValue(tftypes.String, "hello"),
				"computed_attribute":      tftypes.NewValue(tftypes.String, nil),
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"required_nested_attribute": tftypes.String}}, nil),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attributeA": tftypes.String,
						"optional_computed_attributeB": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"optional_computed_attributeA": tftypes.NewValue(tftypes.String, "world"),
					// An unknown in the config represents a situation where
					// an argument is explicitly set to an expression result
					// that is derived from an unknown value. This is distinct
					// from leaving it null, which allows the provider itself
					// to decide the value during PlanResourceChange.
					"optional_computed_attributeB": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				}),
			},
			expectedVal: map[string]tftypes.Value{
				"optional_attribute": tftypes.NewValue(tftypes.String, "hello"),
				// unset computed attributes are null in the proposal; provider
				// usually changes them to "unknown" during PlanResourceChange,
				// to indicate that the value will be decided during apply.
				"computed_attribute":      tftypes.NewValue(tftypes.String, nil),
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"required_nested_attribute": tftypes.String}}, nil),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attributeA": tftypes.String,
						"optional_computed_attributeB": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"optional_computed_attributeA": tftypes.NewValue(tftypes.String, "world"),
					"optional_computed_attributeB": tftypes.NewValue(tftypes.String, tftypes.UnknownValue), // explicit unknown preserved from config
				}),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			request := ProposeNewStateRequest{
				PriorState: tfsdk.State{
					//Raw:    tftypes.NewValue(test.schema.Type().TerraformType(context.Background()), test.priorVal),
					Raw:    tftypes.NewValue(tftypes.DynamicPseudoType, nil),
					Schema: test.schema,
				},
				Config: tfsdk.Config{
					Raw:    tftypes.NewValue(test.schema.Type().TerraformType(context.Background()), test.configVal),
					Schema: test.schema,
				},
			}
			expectedResponse := &ProposeNewStateResponse{
				ProposedNewState: tfsdk.Plan{
					Raw:    tftypes.NewValue(test.schema.Type().TerraformType(context.Background()), test.expectedVal),
					Schema: test.schema,
				},
			}
			response := &ProposeNewStateResponse{}
			SchemaProposeNewState(context.TODO(), test.schema, request, response)
			if diff := cmp.Diff(response, expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
