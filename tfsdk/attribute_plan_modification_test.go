package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attrpath"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseStateForUnknownModifier(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state    attr.Value
		plan     attr.Value
		config   attr.Value
		expected attr.Value
	}

	tests := map[string]testCase{
		"nil-state": {
			// this honestly just shouldn't happen, but let's be
			// sure we're not going to panic if it does
			state:    nil,
			plan:     types.String{Unknown: true},
			config:   types.String{Null: true},
			expected: types.String{Unknown: true},
		},
		"nil-plan": {
			// this honestly just shouldn't happen, but let's be
			// sure we're not going to panic if it does
			state:    types.String{Null: true},
			plan:     nil,
			config:   types.String{Null: true},
			expected: nil,
		},
		"null-state": {
			// when we first create the resource, use the unknown
			// value
			state:    types.String{Null: true},
			plan:     types.String{Unknown: true},
			config:   types.String{Null: true},
			expected: types.String{Unknown: true},
		},
		"known-plan": {
			// this would really only happen if we had a plan
			// modifier setting the value before this plan modifier
			// got to it
			//
			// but we still want to preserve that value, in this
			// case
			state:    types.String{Value: "foo"},
			plan:     types.String{Value: "bar"},
			config:   types.String{Null: true},
			expected: types.String{Value: "bar"},
		},
		"non-null-state-unknown-plan": {
			// this is the situation we want to preserve the state
			// in
			state:    types.String{Value: "foo"},
			plan:     types.String{Unknown: true},
			config:   types.String{Null: true},
			expected: types.String{Value: "foo"},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			state:    types.String{Value: "foo"},
			plan:     types.String{Unknown: true},
			config:   types.String{Unknown: true},
			expected: types.String{Unknown: true},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schema := Schema{
				Attributes: map[string]Attribute{
					"a": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
				},
			}

			var configRaw, planRaw, stateRaw interface{}
			if tc.config != nil {
				val, err := tc.config.ToTerraformValue(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				configRaw = val
			}
			if tc.state != nil {
				val, err := tc.state.ToTerraformValue(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				stateRaw = val
			}
			if tc.plan != nil {
				val, err := tc.plan.ToTerraformValue(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				planRaw = val
			}
			configVal := tftypes.NewValue(tftypes.String, configRaw)
			stateVal := tftypes.NewValue(tftypes.String, stateRaw)
			planVal := tftypes.NewValue(tftypes.String, planRaw)

			req := ModifyAttributePlanRequest{
				AttributePath: attrpath.New(),
				Config: Config{
					Schema: schema,
					Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
						"a": configVal,
					}),
				},
				State: State{
					Schema: schema,
					Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
						"a": stateVal,
					}),
				},
				Plan: Plan{
					Schema: schema,
					Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
						"a": planVal,
					}),
				},
				AttributeConfig: tc.config,
				AttributeState:  tc.state,
				AttributePlan:   tc.plan,
				ProviderMeta:    Config{},
			}
			resp := &ModifyAttributePlanResponse{
				AttributePlan: req.AttributePlan,
			}
			modifier := UseStateForUnknown()

			modifier.Modify(context.Background(), req, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Unexpected diagnostics: %s", resp.Diagnostics)
			}
			if diff := cmp.Diff(tc.expected, resp.AttributePlan); diff != "" {
				t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
			}
		})
	}
}

func TestRequiresReplaceModifier(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state        State
		plan         Plan
		config       Config
		path         attrpath.Path
		expectedPlan attr.Value
		expectedRR   bool
	}

	schema := Schema{
		Attributes: map[string]Attribute{
			"optional-computed": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"optional": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}

	blockSchema := Schema{
		Blocks: map[string]Block{
			"block": {
				Attributes: map[string]Attribute{
					"optional-computed": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"optional": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				NestingMode: BlockNestingModeList,
			},
		},
	}

	tests := map[string]testCase{
		"null-state": {
			// when we first create the resource, it shouldn't
			// require replacing immediately
			state: State{
				Schema: schema,
				Raw:    tftypes.NewValue(schema.TerraformType(context.Background()), nil),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			path:         attrpath.New().Attribute("optional-computed"),
			expectedPlan: types.String{Value: "foo"},
			expectedRR:   false,
		},
		"null-plan": {
			// when we destroy the resource, it shouldn't require
			// replacing
			//
			// Terraform doesn't usually ask for provider input on
			// the plan when destroying resources, but in case it
			// does, let's make sure we handle it right
			plan: Plan{
				Schema: schema,
				Raw:    tftypes.NewValue(schema.TerraformType(context.Background()), nil),
			},
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw:    tftypes.NewValue(schema.TerraformType(context.Background()), nil),
			},
			path:         attrpath.New().Attribute("optional-computed"),
			expectedPlan: nil,
			expectedRR:   false,
		},
		"null-attribute-state": {
			// make sure we're not confusing an attribute going
			// from null to a value with the resource getting
			// created
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			path:         attrpath.New().Attribute("optional"),
			expectedPlan: types.String{Value: "bar"},
			expectedRR:   true,
		},
		"null-attribute-plan": {
			// make sure we're not confusing an attribute going
			// from a value to null with the resource getting
			// destroyed
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			path:         attrpath.New().Attribute("optional"),
			expectedPlan: types.String{Null: true},
			expectedRR:   true,
		},
		"known-state-change": {
			// when updating the attribute, if it has changed, it
			// should require replacing
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			path:         attrpath.New().Attribute("optional"),
			expectedPlan: types.String{Value: "quux"},
			expectedRR:   true,
		},
		"known-state-no-change": {
			// when the attribute hasn't changed, it shouldn't
			// require replacing
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			path:         attrpath.New().Attribute("optional-computed"),
			expectedPlan: types.String{Value: "foo"},
			expectedRR:   false,
		},
		"null-config-computed": {
			// if the config is null for a computed attribute, we
			// shouldn't require replacing, even if it's a change.
			//
			// this is sometimes unintuitive, if the practitioner
			// is changing it on purpose. However, it's
			// indistinguishable from the provider changing it, and
			// practitioners pretty much never expect the resource
			// to be recreated if the provider is the one changing
			// the value.
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, nil),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			path:         attrpath.New().Attribute("optional-computed"),
			expectedPlan: types.String{Unknown: true},
			expectedRR:   false,
		},
		"null-config-not-computed": {
			// if the config is null for a non-computed attribute,
			// we should require replacing if it's a change.
			//
			// unlike computed attributes, this is always a
			// practitioner making a change, and therefore the
			// destroy/recreate cycle is likely expected.
			//
			// this test is technically covered by
			// null-attribute-plan, but let's duplicate it just to
			// be explicit about what each case is actually testing
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			path:         attrpath.New().Attribute("optional"),
			expectedPlan: types.String{Null: true},
			expectedRR:   true,
		},
		"block-no-change": {
			state: State{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
						}),
				}),
			},
			plan: Plan{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
						}),
				}),
			},
			config: Config{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
						}),
				}),
			},
			path: attrpath.New().Attribute("block"),
			expectedPlan: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"optional-computed": types.StringType,
						"optional":          types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"optional-computed": types.StringType,
							"optional":          types.StringType,
						},
						Attrs: map[string]attr.Value{
							"optional-computed": types.String{Value: "samevalue"},
							"optional":          types.String{Value: "samevalue"},
						},
					},
				},
			},
			expectedRR: false,
		},
		"block-element-count-change": {
			state: State{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
						}),
				}),
			},
			plan: Plan{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "newvalue"),
								"optional":          tftypes.NewValue(tftypes.String, "newvalue"),
							}),
						}),
				}),
			},
			config: Config{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "newvalue"),
								"optional":          tftypes.NewValue(tftypes.String, "newvalue"),
							}),
						}),
				}),
			},
			path: attrpath.New().Attribute("block"),
			expectedPlan: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"optional-computed": types.StringType,
						"optional":          types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"optional-computed": types.StringType,
							"optional":          types.StringType,
						},
						Attrs: map[string]attr.Value{
							"optional-computed": types.String{Value: "samevalue"},
							"optional":          types.String{Value: "samevalue"},
						},
					},
					types.Object{
						AttrTypes: map[string]attr.Type{
							"optional-computed": types.StringType,
							"optional":          types.StringType,
						},
						Attrs: map[string]attr.Value{
							"optional-computed": types.String{Value: "newvalue"},
							"optional":          types.String{Value: "newvalue"},
						},
					},
				},
			},
			expectedRR: true,
		},
		"block-nested-attribute-change": {
			state: State{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "oldvalue"),
							}),
						}),
				}),
			},
			plan: Plan{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "newvalue"),
							}),
						}),
				}),
			},
			config: Config{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "newvalue"),
							}),
						}),
				}),
			},
			path: attrpath.New().Attribute("block"),
			expectedPlan: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"optional-computed": types.StringType,
						"optional":          types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"optional-computed": types.StringType,
							"optional":          types.StringType,
						},
						Attrs: map[string]attr.Value{
							"optional-computed": types.String{Value: "samevalue"},
							"optional":          types.String{Value: "newvalue"},
						},
					},
				},
			},
			expectedRR: true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attrConfig, diags := tc.config.getAttributeValue(context.Background(), tc.path)
			if diags.HasError() {
				t.Fatalf("Got unexpected diagnostics: %s", diags)
			}

			attrState, diags := tc.state.getAttributeValue(context.Background(), tc.path)
			if diags.HasError() {
				t.Fatalf("Got unexpected diagnostics: %s", diags)
			}

			attrPlan, diags := tc.plan.getAttributeValue(context.Background(), tc.path)
			if diags.HasError() {
				t.Fatalf("Got unexpected diagnostics: %s", diags)
			}

			req := ModifyAttributePlanRequest{
				AttributePath:   tc.path,
				Config:          tc.config,
				State:           tc.state,
				Plan:            tc.plan,
				AttributeConfig: attrConfig,
				AttributeState:  attrState,
				AttributePlan:   attrPlan,
				ProviderMeta:    Config{},
			}
			resp := &ModifyAttributePlanResponse{
				AttributePlan: req.AttributePlan,
			}
			modifier := RequiresReplace()

			modifier.Modify(context.Background(), req, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Unexpected diagnostics: %s", resp.Diagnostics)
			}
			if diff := cmp.Diff(tc.expectedPlan, resp.AttributePlan); diff != "" {
				t.Fatalf("Unexpected diff in plan (-wanted, +got): %s", diff)
			}
			if diff := cmp.Diff(tc.expectedRR, resp.RequiresReplace); diff != "" {
				t.Fatalf("Unexpected diff in RequiresReplace (-wanted, +got): %s", diff)
			}
		})
	}
}

func TestRequiresReplaceIfModifier(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state        State
		plan         Plan
		config       Config
		priorRR      bool
		path         attrpath.Path
		ifReturn     bool
		expectedPlan attr.Value
		expectedRR   bool
	}

	schema := Schema{
		Attributes: map[string]Attribute{
			"optional-computed": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"optional": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}

	blockSchema := Schema{
		Blocks: map[string]Block{
			"block": {
				Attributes: map[string]Attribute{
					"optional-computed": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"optional": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				NestingMode: BlockNestingModeList,
			},
		},
	}

	tests := map[string]testCase{
		"null-state": {
			// when we first create the resource, it shouldn't
			// require replacing immediately
			state: State{
				Schema: schema,
				Raw:    tftypes.NewValue(schema.TerraformType(context.Background()), nil),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			priorRR:      false,
			path:         attrpath.New().Attribute("optional-computed"),
			ifReturn:     true,
			expectedPlan: types.String{Value: "foo"},
			expectedRR:   false,
		},
		"null-plan": {
			// when we destroy the resource, it shouldn't require
			// replacing
			//
			// Terraform doesn't usually ask for provider input on
			// the plan when destroying resources, but in case it
			// does, let's make sure we handle it right
			plan: Plan{
				Schema: schema,
				Raw:    tftypes.NewValue(schema.TerraformType(context.Background()), nil),
			},
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw:    tftypes.NewValue(schema.TerraformType(context.Background()), nil),
			},
			priorRR:      false,
			path:         attrpath.New().Attribute("optional-computed"),
			ifReturn:     true,
			expectedPlan: nil,
			expectedRR:   false,
		},
		"null-attribute-state": {
			// make sure we're not confusing an attribute going
			// from null to a value with the resource getting
			// created
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			priorRR:      false,
			path:         attrpath.New().Attribute("optional"),
			ifReturn:     true,
			expectedPlan: types.String{Value: "bar"},
			expectedRR:   true,
		},
		"null-attribute-plan": {
			// make sure we're not confusing an attribute going
			// from a value to null with the resource getting
			// destroyed
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			priorRR:      false,
			ifReturn:     true,
			path:         attrpath.New().Attribute("optional"),
			expectedPlan: types.String{Null: true},
			expectedRR:   true,
		},
		"known-state-change-true": {
			// when updating the attribute, if it has changed and
			// the function returns true, it should require
			// replacing
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			priorRR:      false,
			path:         attrpath.New().Attribute("optional"),
			ifReturn:     true,
			expectedPlan: types.String{Value: "quux"},
			expectedRR:   true,
		},
		"known-state-change-false": {
			// when updating the attribute, if it has changed and
			// the function returns false, it should not require
			// replacing
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			priorRR:      false,
			path:         attrpath.New().Attribute("optional"),
			ifReturn:     false,
			expectedPlan: types.String{Value: "quux"},
			expectedRR:   false,
		},
		"known-state-change-false-dont-override": {
			// when updating the attribute, if it has changed and
			// the function returns false, but a prior plan
			// modifier already marked the resource as needing to
			// be recreated, we shouldn't override that
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			priorRR:      true,
			path:         attrpath.New().Attribute("optional"),
			ifReturn:     false,
			expectedPlan: types.String{Value: "quux"},
			expectedRR:   true,
		},
		"known-state-no-change": {
			// when the attribute hasn't changed, it shouldn't
			// require replacing
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			priorRR:      false,
			path:         attrpath.New().Attribute("optional-computed"),
			ifReturn:     true,
			expectedPlan: types.String{Value: "foo"},
			expectedRR:   false,
		},
		"null-config-computed": {
			// if the config is null for a computed attribute, we
			// shouldn't require replacing, even if it's a change.
			//
			// this is sometimes unintuitive, if the practitioner
			// is changing it on purpose. However, it's
			// indistinguishable from the provider changing it, and
			// practitioners pretty much never expect the resource
			// to be recreated if the provider is the one changing
			// the value.
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, nil),
					"optional":          tftypes.NewValue(tftypes.String, "quux"),
				}),
			},
			priorRR:      false,
			path:         attrpath.New().Attribute("optional-computed"),
			ifReturn:     true,
			expectedPlan: types.String{Unknown: true},
			expectedRR:   false,
		},
		"null-config-not-computed": {
			// if the config is null for a non-computed attribute,
			// we should require replacing if it's a change.
			//
			// unlike computed attributes, this is always a
			// practitioner making a change, and therefore the
			// destroy/recreate cycle is likely expected.
			//
			// this test is technically covered by
			// null-attribute-plan, but let's duplicate it just to
			// be explicit about what each case is actually testing
			state: State{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, "bar"),
				}),
			},
			plan: Plan{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			config: Config{
				Schema: schema,
				Raw: tftypes.NewValue(schema.TerraformType(context.Background()), map[string]tftypes.Value{
					"optional-computed": tftypes.NewValue(tftypes.String, "foo"),
					"optional":          tftypes.NewValue(tftypes.String, nil),
				}),
			},
			priorRR:      false,
			path:         attrpath.New().Attribute("optional"),
			ifReturn:     true,
			expectedPlan: types.String{Null: true},
			expectedRR:   true,
		},
		"block-no-change": {
			state: State{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
						}),
				}),
			},
			plan: Plan{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
						}),
				}),
			},
			config: Config{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
						}),
				}),
			},
			path: attrpath.New().Attribute("block"),
			expectedPlan: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"optional-computed": types.StringType,
						"optional":          types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"optional-computed": types.StringType,
							"optional":          types.StringType,
						},
						Attrs: map[string]attr.Value{
							"optional-computed": types.String{Value: "samevalue"},
							"optional":          types.String{Value: "samevalue"},
						},
					},
				},
			},
			ifReturn:   false,
			expectedRR: false,
		},
		"block-element-count-change": {
			state: State{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
						}),
				}),
			},
			plan: Plan{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "newvalue"),
								"optional":          tftypes.NewValue(tftypes.String, "newvalue"),
							}),
						}),
				}),
			},
			config: Config{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "samevalue"),
							}),
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "newvalue"),
								"optional":          tftypes.NewValue(tftypes.String, "newvalue"),
							}),
						}),
				}),
			},
			path: attrpath.New().Attribute("block"),
			expectedPlan: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"optional-computed": types.StringType,
						"optional":          types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"optional-computed": types.StringType,
							"optional":          types.StringType,
						},
						Attrs: map[string]attr.Value{
							"optional-computed": types.String{Value: "samevalue"},
							"optional":          types.String{Value: "samevalue"},
						},
					},
					types.Object{
						AttrTypes: map[string]attr.Type{
							"optional-computed": types.StringType,
							"optional":          types.StringType,
						},
						Attrs: map[string]attr.Value{
							"optional-computed": types.String{Value: "newvalue"},
							"optional":          types.String{Value: "newvalue"},
						},
					},
				},
			},
			ifReturn:   true,
			expectedRR: true,
		},
		"block-nested-attribute-change": {
			state: State{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "oldvalue"),
							}),
						}),
				}),
			},
			plan: Plan{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "newvalue"),
							}),
						}),
				}),
			},
			config: Config{
				Schema: blockSchema,
				Raw: tftypes.NewValue(blockSchema.TerraformType(context.Background()), map[string]tftypes.Value{
					"block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"optional-computed": tftypes.String,
									"optional":          tftypes.String,
								},
							}, map[string]tftypes.Value{
								"optional-computed": tftypes.NewValue(tftypes.String, "samevalue"),
								"optional":          tftypes.NewValue(tftypes.String, "newvalue"),
							}),
						}),
				}),
			},
			path: attrpath.New().Attribute("block"),
			expectedPlan: types.List{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"optional-computed": types.StringType,
						"optional":          types.StringType,
					},
				},
				Elems: []attr.Value{
					types.Object{
						AttrTypes: map[string]attr.Type{
							"optional-computed": types.StringType,
							"optional":          types.StringType,
						},
						Attrs: map[string]attr.Value{
							"optional-computed": types.String{Value: "samevalue"},
							"optional":          types.String{Value: "newvalue"},
						},
					},
				},
			},
			ifReturn:   true,
			expectedRR: true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attrConfig, diags := tc.config.getAttributeValue(context.Background(), tc.path)
			if diags.HasError() {
				t.Fatalf("Got unexpected diagnostics: %s", diags)
			}

			attrState, diags := tc.state.getAttributeValue(context.Background(), tc.path)
			if diags.HasError() {
				t.Fatalf("Got unexpected diagnostics: %s", diags)
			}

			attrPlan, diags := tc.plan.getAttributeValue(context.Background(), tc.path)
			if diags.HasError() {
				t.Fatalf("Got unexpected diagnostics: %s", diags)
			}

			req := ModifyAttributePlanRequest{
				AttributePath:   tc.path,
				Config:          tc.config,
				State:           tc.state,
				Plan:            tc.plan,
				AttributeConfig: attrConfig,
				AttributeState:  attrState,
				AttributePlan:   attrPlan,
				ProviderMeta:    Config{},
			}
			resp := &ModifyAttributePlanResponse{
				AttributePlan:   req.AttributePlan,
				RequiresReplace: tc.priorRR,
			}
			modifier := RequiresReplaceIf(func(ctx context.Context, state, config attr.Value, path attrpath.Path) (bool, diag.Diagnostics) {
				return tc.ifReturn, nil
			}, "", "")

			modifier.Modify(context.Background(), req, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Unexpected diagnostics: %s", resp.Diagnostics)
			}
			if diff := cmp.Diff(tc.expectedPlan, resp.AttributePlan); diff != "" {
				t.Fatalf("Unexpected diff in plan (-wanted, +got): %s", diff)
			}
			if diff := cmp.Diff(tc.expectedRR, resp.RequiresReplace); diff != "" {
				t.Fatalf("Unexpected diff in RequiresReplace (-wanted, +got): %s", diff)
			}
		})
	}
}
