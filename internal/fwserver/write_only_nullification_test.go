// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNullifyWriteOnlyAttributes(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"string-value": schema.StringAttribute{
				Required: true,
			},
			"string-nil": schema.StringAttribute{
				Optional: true,
			},
			"string-nil-write-only": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"string-value-write-only": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"dynamic-value": schema.DynamicAttribute{
				Required: true,
			},
			"dynamic-nil": schema.DynamicAttribute{
				Optional: true,
			},
			"dynamic-nil-write-only": schema.DynamicAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"dynamic-value-write-only": schema.DynamicAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"dynamic-value-with-underlying-list-write-only": schema.DynamicAttribute{
				Optional:  true,
				WriteOnly: true,
			},
			"object-nil-write-only": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"string-nil": types.StringType,
					"string-set": types.StringType,
				},
				Optional:  true,
				WriteOnly: true,
			},
			"object-value-write-only": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"string-nil": types.StringType,
					"string-set": types.StringType,
				},
				Optional:  true,
				WriteOnly: true,
			},
			"nested-nil-write-only": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string-nil": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
					},
					"string-set": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
					},
				},
				Optional:  true,
				WriteOnly: true,
			},
			"nested-value-write-only": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"string-nil": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
					},
					"string-set": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
					},
				},
				Optional:  true,
				WriteOnly: true,
			},
		},
		Blocks: map[string]schema.Block{
			"block-nil-write-only": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string-nil": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
						"string-set": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
					},
				},
			},
			"block-value-write-only": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"string-nil": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
						"string-set": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
					},
				},
			},
		},
	}
	input := tftypes.NewValue(s.Type().TerraformType(context.Background()), map[string]tftypes.Value{
		"string-value":             tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":               tftypes.NewValue(tftypes.String, nil),
		"string-nil-write-only":    tftypes.NewValue(tftypes.String, nil),
		"string-value-write-only":  tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-value":            tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-nil":              tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-nil-write-only":   tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-value-write-only": tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-value-with-underlying-list-write-only": tftypes.NewValue(
			tftypes.List{
				ElementType: tftypes.Bool,
			},
			[]tftypes.Value{
				tftypes.NewValue(tftypes.Bool, true),
				tftypes.NewValue(tftypes.Bool, false),
			},
		),
		"object-nil-write-only": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, nil),
		"object-value-write-only": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "foo"),
		}),
		"nested-nil-write-only": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, nil),
		"nested-value-write-only": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "bar"),
		}),
		"block-nil-write-only": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			},
		}, nil),
		"block-value-write-only": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"string-nil": tftypes.NewValue(tftypes.String, nil),
				"string-set": tftypes.NewValue(tftypes.String, "bar"),
			}),
		}),
	})
	expected := tftypes.NewValue(s.Type().TerraformType(context.Background()), map[string]tftypes.Value{
		"string-value":             tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":               tftypes.NewValue(tftypes.String, nil),
		"string-nil-write-only":    tftypes.NewValue(tftypes.String, nil),
		"string-value-write-only":  tftypes.NewValue(tftypes.String, nil),
		"dynamic-value":            tftypes.NewValue(tftypes.String, "hello, world"),
		"dynamic-nil":              tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-nil-write-only":   tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-value-write-only": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"dynamic-value-with-underlying-list-write-only": tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"object-nil-write-only": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, nil),
		"object-value-write-only": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, nil),
		"nested-nil-write-only": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, nil),
		"nested-value-write-only": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"string-nil": tftypes.String,
				"string-set": tftypes.String,
			},
		}, nil),
		"block-nil-write-only": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			},
		}, nil),
		"block-value-write-only": tftypes.NewValue(tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"string-nil": tftypes.String,
					"string-set": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"string-nil": tftypes.NewValue(tftypes.String, nil),
				"string-set": tftypes.NewValue(tftypes.String, nil),
			}),
		}),
	})

	got, err := tftypes.Transform(input, NullifyWriteOnlyAttributes(context.Background(), s))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	diff, err := expected.Diff(got)
	if err != nil {
		t.Errorf("Error diffing values: %s", err)
		return
	}
	for _, valDiff := range diff {
		t.Errorf("Unexpected diff at path %v: expected: %v, got: %v", valDiff.Path, valDiff.Value1, valDiff.Value2)
	}
}

func TestNullifyWriteOnlyAttributes_NestedTypes(t *testing.T) {
	t.Parallel()
	nestedObjectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"nested-string":    tftypes.String,
			"nested-string-wo": tftypes.String,
		},
	}

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"single-nested-attribute": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"nested-string": schema.StringAttribute{
						Optional: true,
					},
					"nested-string-wo": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
					},
					"nested-single-nested-attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"nested-string": schema.StringAttribute{
								Optional: true,
							},
							"nested-string-wo": schema.StringAttribute{
								Optional:  true,
								WriteOnly: true,
							},
						},
						Optional: true,
					},
					"nested-single-nested-attribute-wo": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"nested-string": schema.StringAttribute{
								Optional: true,
							},
							"nested-string-wo": schema.StringAttribute{
								Optional:  true,
								WriteOnly: true,
							},
						},
						Optional:  true,
						WriteOnly: true,
					},
				},
				Optional: true,
			},
			"single-nested-attribute-wo": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"nested-string": schema.StringAttribute{
						Optional: true,
					},
					"nested-string-wo": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
					},
					"nested-single-nested-attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"nested-string": schema.StringAttribute{
								Optional: true,
							},
							"nested-string-wo": schema.StringAttribute{
								Optional:  true,
								WriteOnly: true,
							},
						},
						Optional: true,
					},
					"nested-single-nested-attribute-wo": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"nested-string": schema.StringAttribute{
								Optional: true,
							},
							"nested-string-wo": schema.StringAttribute{
								Optional:  true,
								WriteOnly: true,
							},
						},
						Optional:  true,
						WriteOnly: true,
					},
				},
				Optional:  true,
				WriteOnly: true,
			},
			"map-nested-attribute": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"nested-string": schema.StringAttribute{
							Optional: true,
						},
						"nested-string-wo": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
						"nested-map-nested-attribute": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
						},
						"nested-map-nested-attribute-wo": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional:  true,
							WriteOnly: true,
						},
					},
				},
				Optional: true,
			},
			"map-nested-attribute-wo": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"nested-string": schema.StringAttribute{
							Optional: true,
						},
						"nested-string-wo": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
						"nested-map-nested-attribute": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional: true,
						},
						"nested-map-nested-attribute-wo": schema.MapNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional:  true,
							WriteOnly: true,
						},
					},
				},
				Optional:  true,
				WriteOnly: true,
			},
			"list-nested-attribute": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"nested-string": schema.StringAttribute{
							Optional: true,
						},
						"nested-string-wo": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
						"nested-list-nested-attribute": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional: true,
						},
						"nested-list-nested-attribute-wo": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional:  true,
							WriteOnly: true,
						},
					},
				},
				Optional: true,
			},
			"list-nested-attribute-wo": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"nested-string": schema.StringAttribute{
							Optional: true,
						},
						"nested-string-wo": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
						"nested-list-nested-attribute": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional: true,
						},
						"nested-list-nested-attribute-wo": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional:  true,
							WriteOnly: true,
						},
					},
				},
				Optional:  true,
				WriteOnly: true,
			},
		},
		Blocks: map[string]schema.Block{
			"single-nested-block": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"nested-string": schema.StringAttribute{
						Optional: true,
					},
					"nested-string-wo": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
					},
					"nested-single-nested-attribute": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"nested-string": schema.StringAttribute{
								Optional: true,
							},
							"nested-string-wo": schema.StringAttribute{
								Optional:  true,
								WriteOnly: true,
							},
						},
						Optional: true,
					},
					"nested-single-nested-attribute-wo": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"nested-string": schema.StringAttribute{
								Optional: true,
							},
							"nested-string-wo": schema.StringAttribute{
								Optional:  true,
								WriteOnly: true,
							},
						},
						Optional:  true,
						WriteOnly: true,
					},
				},
				Blocks: map[string]schema.Block{
					"nested-single-nested-block": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"nested-string": schema.StringAttribute{
								Optional: true,
							},
							"nested-string-wo": schema.StringAttribute{
								Optional:  true,
								WriteOnly: true,
							},
							"nested-single-nested-attribute": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
								Optional: true,
							},
							"nested-single-nested-attribute-wo": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
								Optional:  true,
								WriteOnly: true,
							},
						},
					},
				},
			},
			"list-nested-block": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"nested-string": schema.StringAttribute{
							Optional: true,
						},
						"nested-string-wo": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
						},
						"nested-list-nested-attribute": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional: true,
						},
						"nested-list-nested-attribute-wo": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
							Optional:  true,
							WriteOnly: true,
						},
					},
					Blocks: map[string]schema.Block{
						"nested-list-nested-block": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"nested-string": schema.StringAttribute{
										Optional: true,
									},
									"nested-string-wo": schema.StringAttribute{
										Optional:  true,
										WriteOnly: true,
									},
									"nested-list-nested-attribute": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"nested-string": schema.StringAttribute{
													Optional: true,
												},
												"nested-string-wo": schema.StringAttribute{
													Optional:  true,
													WriteOnly: true,
												},
											},
										},
										Optional: true,
									},
									"nested-list-nested-attribute-wo": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"nested-string": schema.StringAttribute{
													Optional: true,
												},
												"nested-string-wo": schema.StringAttribute{
													Optional:  true,
													WriteOnly: true,
												},
											},
										},
										Optional:  true,
										WriteOnly: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	input := tftypes.NewValue(s.Type().TerraformType(context.Background()), map[string]tftypes.Value{
		"single-nested-attribute": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"nested-string":                     tftypes.String,
				"nested-string-wo":                  tftypes.String,
				"nested-single-nested-attribute":    nestedObjectType,
				"nested-single-nested-attribute-wo": nestedObjectType,
			},
		},
			map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				"nested-single-nested-attribute": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
					"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
					"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				}),
				"nested-single-nested-attribute-wo": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
					"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
					"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				}),
			}),
		"single-nested-attribute-wo": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"nested-string":                     tftypes.String,
				"nested-string-wo":                  tftypes.String,
				"nested-single-nested-attribute":    nestedObjectType,
				"nested-single-nested-attribute-wo": nestedObjectType,
			},
		},
			map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				"nested-single-nested-attribute": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
					"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
					"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				}),
				"nested-single-nested-attribute-wo": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
					"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
					"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				}),
			}),
		"map-nested-attribute": tftypes.NewValue(tftypes.Map{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                  tftypes.String,
					"nested-string-wo":               tftypes.String,
					"nested-map-nested-attribute":    tftypes.Map{ElementType: nestedObjectType},
					"nested-map-nested-attribute-wo": tftypes.Map{ElementType: nestedObjectType},
				},
			},
		}, map[string]tftypes.Value{
			"keyA": tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                  tftypes.String,
					"nested-string-wo":               tftypes.String,
					"nested-map-nested-attribute":    tftypes.Map{ElementType: nestedObjectType},
					"nested-map-nested-attribute-wo": tftypes.Map{ElementType: nestedObjectType},
				},
			}, map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				"nested-map-nested-attribute": tftypes.NewValue(tftypes.Map{ElementType: nestedObjectType}, map[string]tftypes.Value{
					"keyA": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
				"nested-map-nested-attribute-wo": tftypes.NewValue(tftypes.Map{ElementType: nestedObjectType}, map[string]tftypes.Value{
					"keyA": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
			}),
		}),
		"map-nested-attribute-wo": tftypes.NewValue(tftypes.Map{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                  tftypes.String,
					"nested-string-wo":               tftypes.String,
					"nested-map-nested-attribute":    tftypes.Map{ElementType: nestedObjectType},
					"nested-map-nested-attribute-wo": tftypes.Map{ElementType: nestedObjectType},
				},
			},
		}, map[string]tftypes.Value{
			"keyA": tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                  tftypes.String,
					"nested-string-wo":               tftypes.String,
					"nested-map-nested-attribute":    tftypes.Map{ElementType: nestedObjectType},
					"nested-map-nested-attribute-wo": tftypes.Map{ElementType: nestedObjectType},
				},
			}, map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				"nested-map-nested-attribute": tftypes.NewValue(tftypes.Map{ElementType: nestedObjectType}, map[string]tftypes.Value{
					"keyA": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
				"nested-map-nested-attribute-wo": tftypes.NewValue(tftypes.Map{ElementType: nestedObjectType}, map[string]tftypes.Value{
					"keyA": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
			}),
		}),
		"list-nested-attribute": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
				},
			}, map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				"nested-list-nested-attribute": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
					tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
				"nested-list-nested-attribute-wo": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
					tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
			}),
		}),
		"list-nested-attribute-wo": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
				},
			}, map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				"nested-list-nested-attribute": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
					tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
				"nested-list-nested-attribute-wo": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
					tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
			}),
		}),
		"single-nested-block": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"nested-string":                     tftypes.String,
				"nested-string-wo":                  tftypes.String,
				"nested-single-nested-attribute":    nestedObjectType,
				"nested-single-nested-attribute-wo": nestedObjectType,
				"nested-single-nested-block": tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"nested-string":                     tftypes.String,
						"nested-string-wo":                  tftypes.String,
						"nested-single-nested-attribute":    nestedObjectType,
						"nested-single-nested-attribute-wo": nestedObjectType,
					},
				},
			},
		},
			map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				"nested-single-nested-attribute": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
					"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
					"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				}),
				"nested-single-nested-attribute-wo": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
					"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
					"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				}),
				"nested-single-nested-block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"nested-string":                     tftypes.String,
						"nested-string-wo":                  tftypes.String,
						"nested-single-nested-attribute":    nestedObjectType,
						"nested-single-nested-attribute-wo": nestedObjectType,
					},
				},
					map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
						"nested-single-nested-attribute": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
							"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
							"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
						}),
						"nested-single-nested-attribute-wo": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
							"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
							"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
						}),
					}),
			}),
		"list-nested-block": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-block": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested-string":                   tftypes.String,
								"nested-string-wo":                tftypes.String,
								"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
								"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
							},
						},
					},
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-block": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested-string":                   tftypes.String,
								"nested-string-wo":                tftypes.String,
								"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
								"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
							},
						},
					},
				},
			}, map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
				"nested-list-nested-attribute": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
					tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
				"nested-list-nested-attribute-wo": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
					tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
					}),
				}),
				"nested-list-nested-block": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested-string":                   tftypes.String,
							"nested-string-wo":                tftypes.String,
							"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
							"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested-string":                   tftypes.String,
							"nested-string-wo":                tftypes.String,
							"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
							"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
						},
					}, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
						"nested-list-nested-attribute": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
							tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
								"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
								"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
							}),
						}),
						"nested-list-nested-attribute-wo": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
							tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
								"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
								"nested-string-wo": tftypes.NewValue(tftypes.String, "foo-wo"),
							}),
						}),
					}),
				}),
			}),
		}),
	})
	expected := tftypes.NewValue(s.Type().TerraformType(context.Background()), map[string]tftypes.Value{
		"single-nested-attribute": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"nested-string":                     tftypes.String,
				"nested-string-wo":                  tftypes.String,
				"nested-single-nested-attribute":    nestedObjectType,
				"nested-single-nested-attribute-wo": nestedObjectType,
			},
		},
			map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
				"nested-single-nested-attribute": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
					"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
					"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
				}),
				"nested-single-nested-attribute-wo": tftypes.NewValue(nestedObjectType, nil),
			}),
		"single-nested-attribute-wo": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"nested-string":                     tftypes.String,
				"nested-string-wo":                  tftypes.String,
				"nested-single-nested-attribute":    nestedObjectType,
				"nested-single-nested-attribute-wo": nestedObjectType,
			},
		}, nil),
		"map-nested-attribute": tftypes.NewValue(tftypes.Map{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                  tftypes.String,
					"nested-string-wo":               tftypes.String,
					"nested-map-nested-attribute":    tftypes.Map{ElementType: nestedObjectType},
					"nested-map-nested-attribute-wo": tftypes.Map{ElementType: nestedObjectType},
				},
			},
		}, map[string]tftypes.Value{
			"keyA": tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                  tftypes.String,
					"nested-string-wo":               tftypes.String,
					"nested-map-nested-attribute":    tftypes.Map{ElementType: nestedObjectType},
					"nested-map-nested-attribute-wo": tftypes.Map{ElementType: nestedObjectType},
				},
			}, map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
				"nested-map-nested-attribute": tftypes.NewValue(tftypes.Map{ElementType: nestedObjectType}, map[string]tftypes.Value{
					"keyA": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
					}),
				}),
				"nested-map-nested-attribute-wo": tftypes.NewValue(tftypes.Map{ElementType: nestedObjectType}, nil),
			}),
		}),
		"map-nested-attribute-wo": tftypes.NewValue(tftypes.Map{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                  tftypes.String,
					"nested-string-wo":               tftypes.String,
					"nested-map-nested-attribute":    tftypes.Map{ElementType: nestedObjectType},
					"nested-map-nested-attribute-wo": tftypes.Map{ElementType: nestedObjectType},
				},
			},
		}, nil),
		"list-nested-attribute": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
				},
			}, map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
				"nested-list-nested-attribute": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
					tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
					}),
				}),
				"nested-list-nested-attribute-wo": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, nil),
			}),
		}),
		"list-nested-attribute-wo": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
				},
			},
		}, nil),
		"single-nested-block": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"nested-string":                     tftypes.String,
				"nested-string-wo":                  tftypes.String,
				"nested-single-nested-attribute":    nestedObjectType,
				"nested-single-nested-attribute-wo": nestedObjectType,
				"nested-single-nested-block": tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"nested-string":                     tftypes.String,
						"nested-string-wo":                  tftypes.String,
						"nested-single-nested-attribute":    nestedObjectType,
						"nested-single-nested-attribute-wo": nestedObjectType,
					},
				},
			},
		},
			map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
				"nested-single-nested-attribute": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
					"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
					"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
				}),
				"nested-single-nested-attribute-wo": tftypes.NewValue(nestedObjectType, nil),
				"nested-single-nested-block": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"nested-string":                     tftypes.String,
						"nested-string-wo":                  tftypes.String,
						"nested-single-nested-attribute":    nestedObjectType,
						"nested-single-nested-attribute-wo": nestedObjectType,
					},
				},
					map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
						"nested-single-nested-attribute": tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
							"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
							"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
						}),
						"nested-single-nested-attribute-wo": tftypes.NewValue(nestedObjectType, nil),
					}),
			}),
		"list-nested-block": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-block": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested-string":                   tftypes.String,
								"nested-string-wo":                tftypes.String,
								"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
								"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
							},
						},
					},
				},
			},
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"nested-string":                   tftypes.String,
					"nested-string-wo":                tftypes.String,
					"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
					"nested-list-nested-block": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"nested-string":                   tftypes.String,
								"nested-string-wo":                tftypes.String,
								"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
								"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
							},
						},
					},
				},
			}, map[string]tftypes.Value{
				"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
				"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
				"nested-list-nested-attribute": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
					tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
					}),
				}),
				"nested-list-nested-attribute-wo": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, nil),
				"nested-list-nested-block": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested-string":                   tftypes.String,
							"nested-string-wo":                tftypes.String,
							"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
							"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"nested-string":                   tftypes.String,
							"nested-string-wo":                tftypes.String,
							"nested-list-nested-attribute":    tftypes.List{ElementType: nestedObjectType},
							"nested-list-nested-attribute-wo": tftypes.List{ElementType: nestedObjectType},
						},
					}, map[string]tftypes.Value{
						"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
						"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
						"nested-list-nested-attribute": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, []tftypes.Value{
							tftypes.NewValue(nestedObjectType, map[string]tftypes.Value{
								"nested-string":    tftypes.NewValue(tftypes.String, "foo"),
								"nested-string-wo": tftypes.NewValue(tftypes.String, nil),
							}),
						}),
						"nested-list-nested-attribute-wo": tftypes.NewValue(tftypes.List{ElementType: nestedObjectType}, nil),
					}),
				}),
			}),
		}),
	})
	got, err := tftypes.Transform(input, NullifyWriteOnlyAttributes(context.Background(), s))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	diff, err := expected.Diff(got)
	if err != nil {
		t.Errorf("Error diffing values: %s", err)
		return
	}
	for _, valDiff := range diff {
		t.Errorf("Unexpected diff at path %v: expected: %v, got: %v", valDiff.Path, valDiff.Value1, valDiff.Value2)
	}
}
