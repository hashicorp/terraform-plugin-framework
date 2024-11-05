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
		"null block remains null": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"optional_attribute": schema.StringAttribute{
						Optional: true,
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
							"optional_computed_attribute": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
						},
					},
				},
			},
			priorVal: nil,
			configVal: map[string]tftypes.Value{
				"optional_attribute":      tftypes.NewValue(tftypes.String, "bar"),
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"required_nested_attribute": tftypes.String}}, nil),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attribute": tftypes.String,
					},
				}, nil),
			},
			expectedVal: map[string]tftypes.Value{
				"optional_attribute":      tftypes.NewValue(tftypes.String, "bar"),
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"required_nested_attribute": tftypes.String}}, nil),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attribute": tftypes.String,
					},
				}, nil),
			},
		},
		"no prior with set": {
			// This one is here because our handling of sets is more complex
			// than others (due to the fuzzy correlation heuristic) and
			// historically that caused us some panic-related grief.
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						Optional: true,
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"required_nested_attribute": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
				Blocks: map[string]schema.Block{
					"set_nested_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"optional_computed_nested_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			priorVal: nil,
			configVal: map[string]tftypes.Value{
				"set_nested_attribute": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "world"),
						}),
					},
				),
				"set_nested_block": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "blub"),
						}),
					},
				),
			},
			expectedVal: map[string]tftypes.Value{
				"set_nested_attribute": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "world"),
						}),
					},
				),
				"set_nested_block": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "blub"),
						}),
					},
				),
			},
		},
		"prior attributes": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"optional_attribute": schema.StringAttribute{
						Optional: true,
					},
					"computed_attribute": schema.StringAttribute{
						Computed: true,
					},
					"optional_computed_attributeA": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
					"optional_computed_attributeB": schema.StringAttribute{
						Optional: true,
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
			},
			priorVal: map[string]tftypes.Value{
				"optional_attribute":           tftypes.NewValue(tftypes.String, "bonjour"),
				"computed_attribute":           tftypes.NewValue(tftypes.String, "petit dejeuner"),
				"optional_computed_attributeA": tftypes.NewValue(tftypes.String, "grande dejeuner"),
				"optional_computed_attributeB": tftypes.NewValue(tftypes.String, "a la monde"),
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"required_nested_attribute": tftypes.String}},
					map[string]tftypes.Value{
						"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
					}),
			},
			configVal: map[string]tftypes.Value{
				"optional_attribute":           tftypes.NewValue(tftypes.String, "hello"),
				"computed_attribute":           tftypes.NewValue(tftypes.String, nil),
				"optional_computed_attributeA": tftypes.NewValue(tftypes.String, nil),
				"optional_computed_attributeB": tftypes.NewValue(tftypes.String, "world"),
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"required_nested_attribute": tftypes.String}},
					map[string]tftypes.Value{
						"required_nested_attribute": tftypes.NewValue(tftypes.String, "bleep"),
					}),
			},
			expectedVal: map[string]tftypes.Value{
				"optional_attribute":           tftypes.NewValue(tftypes.String, "hello"),
				"computed_attribute":           tftypes.NewValue(tftypes.String, "petit dejeuner"),
				"optional_computed_attributeA": tftypes.NewValue(tftypes.String, "grande dejeuner"),
				"optional_computed_attributeB": tftypes.NewValue(tftypes.String, "world"),
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"required_nested_attribute": tftypes.String}},
					map[string]tftypes.Value{
						"required_nested_attribute": tftypes.NewValue(tftypes.String, "bleep"),
					}),
			},
		},
		"prior nested single": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"required_nested_attribute": schema.StringAttribute{
								Required: true,
							},
							"optional_nested_attribute": schema.StringAttribute{
								Optional: true,
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
			priorVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				},
					map[string]tftypes.Value{
						"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
						"optional_nested_attribute": tftypes.NewValue(tftypes.String, nil),
					}),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attributeA": tftypes.String,
						"optional_computed_attributeB": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"optional_computed_attributeA": tftypes.NewValue(tftypes.String, "bleep"),
					"optional_computed_attributeB": tftypes.NewValue(tftypes.String, "boop"),
				}),
			},
			configVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				},
					map[string]tftypes.Value{
						"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
						"optional_nested_attribute": tftypes.NewValue(tftypes.String, "beep"),
					}),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attributeA": tftypes.String,
						"optional_computed_attributeB": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"optional_computed_attributeA": tftypes.NewValue(tftypes.String, "bap"),
					"optional_computed_attributeB": tftypes.NewValue(tftypes.String, nil),
				}),
			},
			expectedVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				},
					map[string]tftypes.Value{
						"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
						"optional_nested_attribute": tftypes.NewValue(tftypes.String, "beep"),
					}),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attributeA": tftypes.String,
						"optional_computed_attributeB": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"optional_computed_attributeA": tftypes.NewValue(tftypes.String, "bap"),
					"optional_computed_attributeB": tftypes.NewValue(tftypes.String, "boop"),
				}),
			},
		},
		"prior nested single to null": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"required_nested_attribute": schema.StringAttribute{
								Required: true,
							},
							"optional_nested_attribute": schema.StringAttribute{
								Optional: true,
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
			priorVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				},
					map[string]tftypes.Value{
						"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
						"optional_nested_attribute": tftypes.NewValue(tftypes.String, nil),
					}),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attributeA": tftypes.String,
						"optional_computed_attributeB": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"optional_computed_attributeA": tftypes.NewValue(tftypes.String, "bleep"),
					"optional_computed_attributeB": tftypes.NewValue(tftypes.String, "boop"),
				}),
			},
			configVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				}, nil),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attributeA": tftypes.String,
						"optional_computed_attributeB": tftypes.String,
					},
				}, nil),
			},
			expectedVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				}, nil),
				"single_nested_block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"optional_computed_attributeA": tftypes.String,
						"optional_computed_attributeB": tftypes.String,
					},
				}, nil),
			},
		},
		"prior optional computed nested single to null": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"single_nested_attribute": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"required_nested_attribute": schema.StringAttribute{
								Required: true,
							},
							"optional_nested_attribute": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
			priorVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
					"optional_nested_attribute": tftypes.NewValue(tftypes.String, nil),
				}),
			},
			configVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				}, nil),
			},
			expectedVal: map[string]tftypes.Value{
				"single_nested_attribute": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_nested_attribute": tftypes.String,
						"optional_nested_attribute": tftypes.String,
					},
				}, nil),
			},
		},
		"prior nested list": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"required_nested_attribute": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"optional_computed_nested_attributeA": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
								"optional_computed_nested_attributeB": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			priorVal: map[string]tftypes.Value{
				"list_nested_attribute": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "bar"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "baz"),
						}),
					},
				),
				"list_nested_block": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attributeA": tftypes.String,
								"optional_computed_nested_attributeB": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attributeA": tftypes.String,
								"optional_computed_nested_attributeB": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_attributeA": tftypes.NewValue(tftypes.String, "beep"),
							"optional_computed_nested_attributeB": tftypes.NewValue(tftypes.String, "boop"),
						}),
					},
				),
			},
			configVal: map[string]tftypes.Value{
				"list_nested_attribute": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "bar"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "baz"),
						}),
					},
				),
				"list_nested_block": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attributeA": tftypes.String,
								"optional_computed_nested_attributeB": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attributeA": tftypes.String,
								"optional_computed_nested_attributeB": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_attributeA": tftypes.NewValue(tftypes.String, "bap"),
							"optional_computed_nested_attributeB": tftypes.NewValue(tftypes.String, nil),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attributeA": tftypes.String,
								"optional_computed_nested_attributeB": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_attributeA": tftypes.NewValue(tftypes.String, "blep"),
							"optional_computed_nested_attributeB": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
			},
			expectedVal: map[string]tftypes.Value{
				"list_nested_attribute": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "bar"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "baz"),
						}),
					},
				),
				"list_nested_block": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attributeA": tftypes.String,
								"optional_computed_nested_attributeB": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attributeA": tftypes.String,
								"optional_computed_nested_attributeB": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_attributeA": tftypes.NewValue(tftypes.String, "bap"),
							"optional_computed_nested_attributeB": tftypes.NewValue(tftypes.String, "boop"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_attributeA": tftypes.String,
								"optional_computed_nested_attributeB": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_attributeA": tftypes.NewValue(tftypes.String, "blep"),
							"optional_computed_nested_attributeB": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
			},
		},
		"prior nested list with dynamic": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"list_nested_attribute": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"required_nested_dynamic_attributeA": schema.DynamicAttribute{
									Required: true,
								},
								"required_nested_dynamic_attributeB": schema.DynamicAttribute{
									Required: true,
								},
							},
						},
					},
				},
				Blocks: map[string]schema.Block{
					"list_nested_block": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"optional_computed_nested_string_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
								"optional_computed_nested_dynamic_attribute": schema.DynamicAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			priorVal: map[string]tftypes.Value{
				"list_nested_attribute": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_dynamic_attributeA": tftypes.DynamicPseudoType,
								"required_nested_dynamic_attributeB": tftypes.DynamicPseudoType,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_dynamic_attributeA": tftypes.DynamicPseudoType,
								"required_nested_dynamic_attributeB": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_dynamic_attributeA": tftypes.NewValue(tftypes.String, "bar"),
							"required_nested_dynamic_attributeB": tftypes.NewValue(tftypes.String, "glup"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_dynamic_attributeA": tftypes.DynamicPseudoType,
								"required_nested_dynamic_attributeB": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_dynamic_attributeA": tftypes.NewValue(tftypes.String, "baz"),
							"required_nested_dynamic_attributeB": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
				"list_nested_block": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_string_attribute":  tftypes.String,
								"optional_computed_nested_dynamic_attribute": tftypes.DynamicPseudoType,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_string_attribute":  tftypes.String,
								"optional_computed_nested_dynamic_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_string_attribute":  tftypes.NewValue(tftypes.String, "beep"),
							"optional_computed_nested_dynamic_attribute": tftypes.NewValue(tftypes.String, "boop"),
						}),
					},
				),
			},
			configVal: map[string]tftypes.Value{
				"list_nested_attribute": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_dynamic_attributeA": tftypes.DynamicPseudoType,
								"required_nested_dynamic_attributeB": tftypes.DynamicPseudoType,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_dynamic_attributeA": tftypes.DynamicPseudoType,
								"required_nested_dynamic_attributeB": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_dynamic_attributeA": tftypes.NewValue(tftypes.String, "bar"),
							"required_nested_dynamic_attributeB": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
				"list_nested_block": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_string_attribute":  tftypes.String,
								"optional_computed_nested_dynamic_attribute": tftypes.DynamicPseudoType,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_string_attribute":  tftypes.String,
								"optional_computed_nested_dynamic_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_string_attribute":  tftypes.NewValue(tftypes.String, "bap"),
							"optional_computed_nested_dynamic_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_string_attribute":  tftypes.String,
								"optional_computed_nested_dynamic_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_string_attribute":  tftypes.NewValue(tftypes.String, "blep"),
							"optional_computed_nested_dynamic_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
			},
			expectedVal: map[string]tftypes.Value{
				"list_nested_attribute": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_dynamic_attributeA": tftypes.DynamicPseudoType,
								"required_nested_dynamic_attributeB": tftypes.DynamicPseudoType,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_dynamic_attributeA": tftypes.DynamicPseudoType,
								"required_nested_dynamic_attributeB": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_dynamic_attributeA": tftypes.NewValue(tftypes.String, "bar"),
							"required_nested_dynamic_attributeB": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
				"list_nested_block": tftypes.NewValue(
					tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_string_attribute":  tftypes.String,
								"optional_computed_nested_dynamic_attribute": tftypes.DynamicPseudoType,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_string_attribute":  tftypes.String,
								"optional_computed_nested_dynamic_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_string_attribute":  tftypes.NewValue(tftypes.String, "bap"),
							"optional_computed_nested_dynamic_attribute": tftypes.NewValue(tftypes.String, "boop"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_computed_nested_string_attribute":  tftypes.String,
								"optional_computed_nested_dynamic_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"optional_computed_nested_string_attribute":  tftypes.NewValue(tftypes.String, "blep"),
							"optional_computed_nested_dynamic_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
			},
		},
		"prior nested map": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"required_nested_attribute": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
			},
			priorVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
						}),
						"b": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "blub"),
						}),
					},
				),
			},
			configVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
						}),
						"c": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "blub"),
						}),
					},
				),
			},
			expectedVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
						}),
						"c": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "blub"),
						}),
					},
				),
			},
		},
		"prior optional computed nested map elem to null": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"optional_nested_attribute": schema.StringAttribute{
									Optional: true,
								},
								"optional_computed_nested_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			priorVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "glub"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "computed"),
						}),
						"b": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "blub"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "computed"),
						}),
					},
				),
			},
			configVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, nil),
						"c": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "blub"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
			},
			expectedVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, nil),
						"c": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "blub"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
			},
		},
		"prior optional computed nested map to null": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						Optional: true,
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"optional_nested_attribute": schema.StringAttribute{
									Optional: true,
								},
								"optional_computed_nested_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			priorVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "glub"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "computed"),
						}),
						"b": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "blub"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "computed"),
						}),
					},
				),
			},
			configVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					nil,
				),
			},
			expectedVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					nil,
				),
			},
		},
		"prior nested map with dynamic": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"map_nested_attribute": schema.MapNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"required_nested_attribute": schema.DynamicAttribute{
									Required: true,
								},
							},
						},
					},
				},
			},
			priorVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
						}),
						"b": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.Number, 13),
						}),
					},
				),
			},
			configVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "blep"),
						}),
						"c": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.Number, 13),
						}),
					},
				),
			},
			expectedVal: map[string]tftypes.Value{
				"map_nested_attribute": tftypes.NewValue(
					tftypes.Map{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						},
					},
					map[string]tftypes.Value{
						"a": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "blep"),
						}),
						"c": tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.DynamicPseudoType,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.Number, 13),
						}),
					},
				),
			},
		},
		"prior nested set": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"set_nested_attribute": schema.SetNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"required_nested_attribute": schema.StringAttribute{
									Required: true,
								},
								"optional_nested_attribute": schema.StringAttribute{
									Optional: true,
								},
							},
						},
					},
				},
				Blocks: map[string]schema.Block{
					"set_nested_block": schema.SetNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								// This non-computed attribute will serve
								// as our matching key for propagating
								// "optional_computed_nested_attribute" from elements in the prior value.
								"optional_nested_attribute": schema.StringAttribute{
									Optional: true,
								},
								"optional_computed_nested_attribute": schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
			priorVal: map[string]tftypes.Value{
				"set_nested_attribute": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glubglub"),
							"optional_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glubglub"),
							"optional_nested_attribute": tftypes.NewValue(tftypes.String, "beep"),
						}),
					},
				),
				"set_nested_block": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "beep"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "boop"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "blep"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "boot"),
						}),
					},
				),
			},
			configVal: map[string]tftypes.Value{
				"set_nested_attribute": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glubglub"),
							"optional_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
							"optional_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
				"set_nested_block": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "beep"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "bosh"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
			},
			expectedVal: map[string]tftypes.Value{
				"set_nested_attribute": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glubglub"),
							"optional_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_nested_attribute": tftypes.String,
								"optional_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_nested_attribute": tftypes.NewValue(tftypes.String, "glub"),
							"optional_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
				"set_nested_block": tftypes.NewValue(
					tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						},
					},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "beep"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, "boop"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"optional_nested_attribute":          tftypes.String,
								"optional_computed_nested_attribute": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"optional_nested_attribute":          tftypes.NewValue(tftypes.String, "bosh"),
							"optional_computed_nested_attribute": tftypes.NewValue(tftypes.String, nil),
						}),
					},
				),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			priorStateVal := tftypes.NewValue(tftypes.DynamicPseudoType, nil)
			if test.priorVal != nil {
				priorStateVal = tftypes.NewValue(test.schema.Type().TerraformType(context.Background()), test.priorVal)
			}

			request := ProposeNewStateRequest{
				PriorState: tfsdk.State{
					Raw:    priorStateVal,
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
