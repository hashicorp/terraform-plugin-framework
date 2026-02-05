// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestResourceSchema(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input       *tfprotov6.Schema
		expected    *resourceschema.Schema
		expectedErr string
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"no-block": {
			input:    &tfprotov6.Schema{},
			expected: nil,
		},
		"no-attrs-no-nested-blocks": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{},
			},
			expected: &resourceschema.Schema{
				Attributes: make(map[string]resourceschema.Attribute, 0),
				Blocks:     make(map[string]resourceschema.Block, 0),
			},
		},
		"primitives-attrs": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bool",
							Type:     tftypes.Bool,
							Required: true,
						},
						{
							Name:     "number",
							Type:     tftypes.Number,
							Optional: true,
							Computed: true,
						},
						{
							Name:      "string",
							Type:      tftypes.String,
							Optional:  true,
							Computed:  true,
							Sensitive: true,
						},
						{
							Name:      "dynamic",
							Type:      tftypes.DynamicPseudoType,
							Optional:  true,
							WriteOnly: true,
						},
					},
				},
			},
			expected: &resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"bool": resourceschema.BoolAttribute{
						Required: true,
					},
					"number": resourceschema.NumberAttribute{
						Optional: true,
						Computed: true,
					},
					"string": resourceschema.StringAttribute{
						Optional:  true,
						Computed:  true,
						Sensitive: true,
					},
					"dynamic": resourceschema.DynamicAttribute{
						Optional:  true,
						WriteOnly: true,
					},
				},
				Blocks: make(map[string]resourceschema.Block, 0),
			},
		},
		"collection-attrs": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "list_of_bools",
							Type:      tftypes.List{ElementType: tftypes.Bool},
							Required:  true,
							WriteOnly: true,
						},
						{
							Name:     "map_of_numbers",
							Type:     tftypes.Map{ElementType: tftypes.Number},
							Optional: true,
							Computed: true,
						},
						{
							Name:      "set_of_strings",
							Type:      tftypes.Set{ElementType: tftypes.String},
							Optional:  true,
							Computed:  true,
							Sensitive: true,
						},
						{
							Name: "list_of_objects",
							Type: tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"dynamic": tftypes.DynamicPseudoType,
										"string":  tftypes.String,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			expected: &resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"list_of_bools": resourceschema.ListAttribute{
						ElementType: basetypes.BoolType{},
						Required:    true,
						WriteOnly:   true,
					},
					"map_of_numbers": resourceschema.MapAttribute{
						ElementType: basetypes.NumberType{},
						Optional:    true,
						Computed:    true,
					},
					"set_of_strings": resourceschema.SetAttribute{
						ElementType: basetypes.StringType{},
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
					},
					"list_of_objects": resourceschema.ListAttribute{
						ElementType: basetypes.ObjectType{
							AttrTypes: map[string]attr.Type{
								"dynamic": basetypes.DynamicType{},
								"string":  basetypes.StringType{},
							},
						},
						Required: true,
					},
				},
				Blocks: make(map[string]resourceschema.Block, 0),
			},
		},
		"object-attr": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name: "object",
							Type: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"bool":    tftypes.Bool,
									"dynamic": tftypes.DynamicPseudoType,
									"string":  tftypes.String,
								},
							},
							Optional:  true,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
			},
			expected: &resourceschema.Schema{
				Attributes: map[string]resourceschema.Attribute{
					"object": resourceschema.ObjectAttribute{
						AttributeTypes: map[string]attr.Type{
							"bool":    basetypes.BoolType{},
							"dynamic": basetypes.DynamicType{},
							"string":  basetypes.StringType{},
						},
						Optional:  true,
						Computed:  true,
						Sensitive: true,
					},
				},
				Blocks: make(map[string]resourceschema.Block, 0),
			},
		},
		"tuple-error": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name: "tuple",
							Type: tftypes.Tuple{
								ElementTypes: []tftypes.Type{
									tftypes.Bool,
									tftypes.Number,
									tftypes.String,
								},
							},
							Required: true,
						},
					},
				},
			},
			expectedErr: `no supported attribute for "tuple", type: tftypes.Tuple`,
		},
		"list-nested": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "list_nested",
							Required: true,
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeList,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "list_of_strings",
										Type:     tftypes.List{ElementType: tftypes.String},
										Computed: true,
									},
									{
										Name:     "nested_list_attr",
										Required: true,
										NestedType: &tfprotov6.SchemaObject{
											Nesting: tfprotov6.SchemaObjectNestingModeList,
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:      "bool",
													Type:      tftypes.Bool,
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
				},
			},
			expected: &resourceschema.Schema{
				Blocks: make(map[string]resourceschema.Block, 0),
				Attributes: map[string]resourceschema.Attribute{
					"list_nested": resourceschema.ListNestedAttribute{
						Required: true,
						NestedObject: resourceschema.NestedAttributeObject{
							Attributes: map[string]resourceschema.Attribute{
								"list_of_strings": resourceschema.ListAttribute{
									ElementType: basetypes.StringType{},
									Computed:    true,
								},
								"nested_list_attr": resourceschema.ListNestedAttribute{
									Required: true,
									NestedObject: resourceschema.NestedAttributeObject{
										Attributes: map[string]resourceschema.Attribute{
											"bool": resourceschema.BoolAttribute{
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
			},
		},
		"list-block": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							TypeName: "list_block",
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "list_of_strings",
										Type:     tftypes.List{ElementType: tftypes.String},
										Computed: true,
									},
								},
								BlockTypes: []*tfprotov6.SchemaNestedBlock{
									{
										TypeName: "nested_list_block",
										Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
										Block: &tfprotov6.SchemaBlock{
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:     "bool",
													Type:     tftypes.Bool,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &resourceschema.Schema{
				Blocks: map[string]resourceschema.Block{
					"list_block": resourceschema.ListNestedBlock{
						NestedObject: resourceschema.NestedBlockObject{
							Attributes: map[string]resourceschema.Attribute{
								"list_of_strings": resourceschema.ListAttribute{
									ElementType: basetypes.StringType{},
									Computed:    true,
								},
							},
							Blocks: map[string]resourceschema.Block{
								"nested_list_block": resourceschema.ListNestedBlock{
									NestedObject: resourceschema.NestedBlockObject{
										Attributes: map[string]resourceschema.Attribute{
											"bool": resourceschema.BoolAttribute{
												Required: true,
											},
										},
									},
								},
							},
						},
					},
				},
				Attributes: make(map[string]resourceschema.Attribute, 0),
			},
		},
		"set-nested": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "set_nested",
							Required: true,
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeSet,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "set_of_strings",
										Type:     tftypes.Set{ElementType: tftypes.String},
										Computed: true,
									},
									{
										Name:     "nested_set_attr",
										Required: true,
										NestedType: &tfprotov6.SchemaObject{
											Nesting: tfprotov6.SchemaObjectNestingModeSet,
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:     "bool",
													Type:     tftypes.Bool,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &resourceschema.Schema{
				Blocks: make(map[string]resourceschema.Block, 0),
				Attributes: map[string]resourceschema.Attribute{
					"set_nested": resourceschema.SetNestedAttribute{
						Required: true,
						NestedObject: resourceschema.NestedAttributeObject{
							Attributes: map[string]resourceschema.Attribute{
								"set_of_strings": resourceschema.SetAttribute{
									ElementType: basetypes.StringType{},
									Computed:    true,
								},
								"nested_set_attr": resourceschema.SetNestedAttribute{
									Required: true,
									NestedObject: resourceschema.NestedAttributeObject{
										Attributes: map[string]resourceschema.Attribute{
											"bool": resourceschema.BoolAttribute{
												Optional: true,
												Computed: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"set-block": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							TypeName: "set_block",
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "set_of_strings",
										Type:     tftypes.Set{ElementType: tftypes.String},
										Computed: true,
									},
								},
								BlockTypes: []*tfprotov6.SchemaNestedBlock{
									{
										TypeName: "nested_set_block",
										Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
										Block: &tfprotov6.SchemaBlock{
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:     "bool",
													Type:     tftypes.Bool,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &resourceschema.Schema{
				Blocks: map[string]resourceschema.Block{
					"set_block": resourceschema.SetNestedBlock{
						NestedObject: resourceschema.NestedBlockObject{
							Attributes: map[string]resourceschema.Attribute{
								"set_of_strings": resourceschema.SetAttribute{
									ElementType: basetypes.StringType{},
									Computed:    true,
								},
							},
							Blocks: map[string]resourceschema.Block{
								"nested_set_block": resourceschema.SetNestedBlock{
									NestedObject: resourceschema.NestedBlockObject{
										Attributes: map[string]resourceschema.Attribute{
											"bool": resourceschema.BoolAttribute{
												Required: true,
											},
										},
									},
								},
							},
						},
					},
				},
				Attributes: make(map[string]resourceschema.Attribute, 0),
			},
		},
		"single-nested": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "single_nested",
							Required: true,
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeSingle,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "dynamic",
										Type:     tftypes.DynamicPseudoType,
										Computed: true,
									},
									{
										Name:     "nested_single_attr",
										Required: true,
										NestedType: &tfprotov6.SchemaObject{
											Nesting: tfprotov6.SchemaObjectNestingModeSingle,
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:      "bool",
													Type:      tftypes.Bool,
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
				},
			},
			expected: &resourceschema.Schema{
				Blocks: make(map[string]resourceschema.Block, 0),
				Attributes: map[string]resourceschema.Attribute{
					"single_nested": resourceschema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]resourceschema.Attribute{
							"dynamic": resourceschema.DynamicAttribute{
								Computed: true,
							},
							"nested_single_attr": resourceschema.SingleNestedAttribute{
								Required: true,
								Attributes: map[string]resourceschema.Attribute{
									"bool": resourceschema.BoolAttribute{
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
		"single-block": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							TypeName: "single_block",
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "map_of_strings",
										Type:     tftypes.Map{ElementType: tftypes.String},
										Computed: true,
									},
								},
								BlockTypes: []*tfprotov6.SchemaNestedBlock{
									{
										TypeName: "nested_single_block",
										Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
										Block: &tfprotov6.SchemaBlock{
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:     "bool",
													Type:     tftypes.Bool,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &resourceschema.Schema{
				Blocks: map[string]resourceschema.Block{
					"single_block": resourceschema.SingleNestedBlock{
						Attributes: map[string]resourceschema.Attribute{
							"map_of_strings": resourceschema.MapAttribute{
								ElementType: basetypes.StringType{},
								Computed:    true,
							},
						},
						Blocks: map[string]resourceschema.Block{
							"nested_single_block": resourceschema.SingleNestedBlock{
								Attributes: map[string]resourceschema.Attribute{
									"bool": resourceschema.BoolAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
				Attributes: make(map[string]resourceschema.Attribute, 0),
			},
		},
		"map-nested": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "map_nested",
							Required: true,
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeMap,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "map_of_strings",
										Type:     tftypes.Map{ElementType: tftypes.String},
										Computed: true,
									},
									{
										Name:     "nested_map_attr",
										Required: true,
										NestedType: &tfprotov6.SchemaObject{
											Nesting: tfprotov6.SchemaObjectNestingModeMap,
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:      "bool",
													Type:      tftypes.Bool,
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
				},
			},
			expected: &resourceschema.Schema{
				Blocks: make(map[string]resourceschema.Block, 0),
				Attributes: map[string]resourceschema.Attribute{
					"map_nested": resourceschema.MapNestedAttribute{
						Required: true,
						NestedObject: resourceschema.NestedAttributeObject{
							Attributes: map[string]resourceschema.Attribute{
								"map_of_strings": resourceschema.MapAttribute{
									ElementType: basetypes.StringType{},
									Computed:    true,
								},
								"nested_map_attr": resourceschema.MapNestedAttribute{
									Required: true,
									NestedObject: resourceschema.NestedAttributeObject{
										Attributes: map[string]resourceschema.Attribute{
											"bool": resourceschema.BoolAttribute{
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
			},
		},
		"map-block": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							TypeName: "map_block",
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeMap,
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			expectedErr: `no supported block for nesting mode MAP in nested block "map_block"`,
		},
		"block-with-nested-attr": {
			input: &tfprotov6.Schema{
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							TypeName: "single_block",
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "list_nested",
										Required: true,
										NestedType: &tfprotov6.SchemaObject{
											Nesting: tfprotov6.SchemaObjectNestingModeList,
											Attributes: []*tfprotov6.SchemaAttribute{
												{
													Name:     "number",
													Type:     tftypes.Number,
													Computed: true,
												},
												{
													Name:     "nested_map_attr",
													Required: true,
													NestedType: &tfprotov6.SchemaObject{
														Nesting: tfprotov6.SchemaObjectNestingModeMap,
														Attributes: []*tfprotov6.SchemaAttribute{
															{
																Name:      "bool",
																Type:      tftypes.Bool,
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
							},
						},
					},
				},
			},
			expected: &resourceschema.Schema{
				Blocks: map[string]resourceschema.Block{
					"single_block": resourceschema.SingleNestedBlock{
						Attributes: map[string]resourceschema.Attribute{
							"list_nested": resourceschema.ListNestedAttribute{
								Required: true,
								NestedObject: resourceschema.NestedAttributeObject{
									Attributes: map[string]resourceschema.Attribute{
										"number": resourceschema.NumberAttribute{
											Computed: true,
										},
										"nested_map_attr": resourceschema.MapNestedAttribute{
											Required: true,
											NestedObject: resourceschema.NestedAttributeObject{
												Attributes: map[string]resourceschema.Attribute{
													"bool": resourceschema.BoolAttribute{
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
					},
				},
				Attributes: make(map[string]resourceschema.Attribute, 0),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := fromproto6.ResourceSchema(context.Background(), tc.input)
			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if err.Error() != tc.expectedErr {
					t.Errorf("Expected error to be %q, got %q", tc.expectedErr, err.Error())
					return
				}
				// got expected error
				return
			}
			if tc.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", tc.expectedErr)
				return
			}
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}
