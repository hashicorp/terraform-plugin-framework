// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package fromproto5_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto5"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestResourceSchema(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input       *tfprotov5.Schema
		expected    *resourceschema.Schema
		expectedErr string
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"no-block": {
			input:    &tfprotov5.Schema{},
			expected: nil,
		},
		"no-attrs-no-nested-blocks": {
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{},
			},
			expected: &resourceschema.Schema{
				Attributes: make(map[string]resourceschema.Attribute, 0),
				Blocks:     make(map[string]resourceschema.Block, 0),
			},
		},
		"primitives-attrs": {
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
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
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
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
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
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
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
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
		"list-block": {
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					BlockTypes: []*tfprotov5.SchemaNestedBlock{
						{
							TypeName: "list_block",
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
							Block: &tfprotov5.SchemaBlock{
								Attributes: []*tfprotov5.SchemaAttribute{
									{
										Name:     "list_of_strings",
										Type:     tftypes.List{ElementType: tftypes.String},
										Computed: true,
									},
								},
								BlockTypes: []*tfprotov5.SchemaNestedBlock{
									{
										TypeName: "nested_list_block",
										Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
										Block: &tfprotov5.SchemaBlock{
											Attributes: []*tfprotov5.SchemaAttribute{
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
		"set-block": {
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					BlockTypes: []*tfprotov5.SchemaNestedBlock{
						{
							TypeName: "set_block",
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
							Block: &tfprotov5.SchemaBlock{
								Attributes: []*tfprotov5.SchemaAttribute{
									{
										Name:     "set_of_strings",
										Type:     tftypes.Set{ElementType: tftypes.String},
										Computed: true,
									},
								},
								BlockTypes: []*tfprotov5.SchemaNestedBlock{
									{
										TypeName: "nested_set_block",
										Nesting:  tfprotov5.SchemaNestedBlockNestingModeSet,
										Block: &tfprotov5.SchemaBlock{
											Attributes: []*tfprotov5.SchemaAttribute{
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
		"single-block": {
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					BlockTypes: []*tfprotov5.SchemaNestedBlock{
						{
							TypeName: "single_block",
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeSingle,
							Block: &tfprotov5.SchemaBlock{
								Attributes: []*tfprotov5.SchemaAttribute{
									{
										Name:     "map_of_strings",
										Type:     tftypes.Map{ElementType: tftypes.String},
										Computed: true,
									},
								},
								BlockTypes: []*tfprotov5.SchemaNestedBlock{
									{
										TypeName: "nested_single_block",
										Nesting:  tfprotov5.SchemaNestedBlockNestingModeSingle,
										Block: &tfprotov5.SchemaBlock{
											Attributes: []*tfprotov5.SchemaAttribute{
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
		"map-block": {
			input: &tfprotov5.Schema{
				Block: &tfprotov5.SchemaBlock{
					BlockTypes: []*tfprotov5.SchemaNestedBlock{
						{
							TypeName: "map_block",
							Nesting:  tfprotov5.SchemaNestedBlockNestingModeMap,
							Block: &tfprotov5.SchemaBlock{
								Attributes: []*tfprotov5.SchemaAttribute{
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
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := fromproto5.ResourceSchema(context.Background(), tc.input)
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
