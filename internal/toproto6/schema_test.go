// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSchema(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       fwschema.Schema
		expected    *tfprotov6.Schema
		expectedErr string
	}

	tests := map[string]testCase{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty-val": {
			input: testschema.Schema{},
			expected: &tfprotov6.Schema{
				Block:   &tfprotov6.SchemaBlock{},
				Version: 0,
			},
		},
		"basic-attrs": {
			input: testschema.Schema{
				Version: 1,
				Attributes: map[string]fwschema.Attribute{
					"string": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
					"number": testschema.Attribute{
						Type:     types.NumberType,
						Optional: true,
					},
					"bool": testschema.Attribute{
						Type:     types.BoolType,
						Computed: true,
					},
				},
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bool",
							Type:     tftypes.Bool,
							Computed: true,
						},
						{
							Name:     "number",
							Type:     tftypes.Number,
							Optional: true,
						},
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
				},
			},
		},
		"complex-attrs": {
			input: testschema.Schema{
				Version: 2,
				Attributes: map[string]fwschema.Attribute{
					"list": testschema.Attribute{
						Type:     types.ListType{ElemType: types.StringType},
						Required: true,
					},
					"object": testschema.Attribute{
						Type: types.ObjectType{AttrTypes: map[string]attr.Type{
							"string": types.StringType,
							"number": types.NumberType,
							"bool":   types.BoolType,
						}},
						Optional: true,
					},
					"map": testschema.Attribute{
						Type:     types.MapType{ElemType: types.NumberType},
						Computed: true,
					},
					"set": testschema.Attribute{
						Type:     types.SetType{ElemType: types.StringType},
						Required: true,
					},
					// TODO: add tuple support when it lands
				},
			},
			expected: &tfprotov6.Schema{
				Version: 2,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "list",
							Type:     tftypes.List{ElementType: tftypes.String},
							Required: true,
						},
						{
							Name:     "map",
							Type:     tftypes.Map{ElementType: tftypes.Number},
							Computed: true,
						},
						{
							Name: "object",
							Type: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
								"string": tftypes.String,
								"number": tftypes.Number,
								"bool":   tftypes.Bool,
							}},
							Optional: true,
						},
						{
							Name:     "set",
							Type:     tftypes.Set{ElementType: tftypes.String},
							Required: true,
						},
					},
				},
			},
		},
		"nested-attrs": {
			input: testschema.Schema{
				Version: 3,
				Attributes: map[string]fwschema.Attribute{
					"single": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"string": testschema.Attribute{
									Type:     types.StringType,
									Required: true,
								},
								"number": testschema.Attribute{
									Type:     types.NumberType,
									Optional: true,
								},
								"bool": testschema.Attribute{
									Type:     types.BoolType,
									Computed: true,
								},
								"list": testschema.Attribute{
									Type:     types.ListType{ElemType: types.StringType},
									Computed: true,
									Optional: true,
								},
							},
						},
						NestingMode: fwschema.NestingModeSingle,
						Required:    true,
					},
					"list": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"string": testschema.Attribute{
									Type:     types.StringType,
									Required: true,
								},
								"number": testschema.Attribute{
									Type:     types.NumberType,
									Optional: true,
								},
								"bool": testschema.Attribute{
									Type:     types.BoolType,
									Computed: true,
								},
								"list": testschema.Attribute{
									Type:     types.ListType{ElemType: types.StringType},
									Computed: true,
									Optional: true,
								},
							},
						},
						NestingMode: fwschema.NestingModeList,
						Optional:    true,
					},
					"set": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"string": testschema.Attribute{
									Type:     types.StringType,
									Required: true,
								},
								"number": testschema.Attribute{
									Type:     types.NumberType,
									Optional: true,
								},
								"bool": testschema.Attribute{
									Type:     types.BoolType,
									Computed: true,
								},
								"list": testschema.Attribute{
									Type:     types.ListType{ElemType: types.StringType},
									Computed: true,
									Optional: true,
								},
							},
						},
						NestingMode: fwschema.NestingModeSet,
						Computed:    true,
					},
					"map": testschema.NestedAttribute{
						NestedObject: testschema.NestedAttributeObject{
							Attributes: map[string]fwschema.Attribute{
								"string": testschema.Attribute{
									Type:     types.StringType,
									Required: true,
								},
								"number": testschema.Attribute{
									Type:     types.NumberType,
									Optional: true,
								},
								"bool": testschema.Attribute{
									Type:     types.BoolType,
									Computed: true,
								},
								"list": testschema.Attribute{
									Type:     types.ListType{ElemType: types.StringType},
									Computed: true,
									Optional: true,
								},
							},
						},
						NestingMode: fwschema.NestingModeMap,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			expected: &tfprotov6.Schema{
				Version: 3,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name: "list",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeList,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Optional: true,
						},
						{
							Name: "map",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeMap,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Optional: true,
							Computed: true,
						},
						{
							Name: "set",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeSet,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Computed: true,
						},
						{
							Name: "single",
							NestedType: &tfprotov6.SchemaObject{
								Nesting: tfprotov6.SchemaObjectNestingModeSingle,
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "bool",
										Type:     tftypes.Bool,
										Computed: true,
									},
									{
										Name:     "list",
										Type:     tftypes.List{ElementType: tftypes.String},
										Optional: true,
										Computed: true,
									},
									{
										Name:     "number",
										Type:     tftypes.Number,
										Optional: true,
									},
									{
										Name:     "string",
										Type:     tftypes.String,
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
		},
		"nested-blocks": {
			input: testschema.Schema{
				Version: 3,
				Blocks: map[string]fwschema.Block{
					"list": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"string": testschema.Attribute{
									Type:     types.StringType,
									Required: true,
								},
								"number": testschema.Attribute{
									Type:     types.NumberType,
									Optional: true,
								},
								"bool": testschema.Attribute{
									Type:     types.BoolType,
									Computed: true,
								},
								"list": testschema.Attribute{
									Type:     types.ListType{ElemType: types.StringType},
									Computed: true,
									Optional: true,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeList,
					},
					"set": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"string": testschema.Attribute{
									Type:     types.StringType,
									Required: true,
								},
								"number": testschema.Attribute{
									Type:     types.NumberType,
									Optional: true,
								},
								"bool": testschema.Attribute{
									Type:     types.BoolType,
									Computed: true,
								},
								"list": testschema.Attribute{
									Type:     types.ListType{ElemType: types.StringType},
									Computed: true,
									Optional: true,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeSet,
					},
					"single": testschema.Block{
						NestedObject: testschema.NestedBlockObject{
							Attributes: map[string]fwschema.Attribute{
								"string": testschema.Attribute{
									Type:     types.StringType,
									Required: true,
								},
								"number": testschema.Attribute{
									Type:     types.NumberType,
									Optional: true,
								},
								"bool": testschema.Attribute{
									Type:     types.BoolType,
									Computed: true,
								},
								"list": testschema.Attribute{
									Type:     types.ListType{ElemType: types.StringType},
									Computed: true,
									Optional: true,
								},
							},
						},
						NestingMode: fwschema.BlockNestingModeSingle,
					},
				},
			},
			expected: &tfprotov6.Schema{
				Version: 3,
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Computed: true,
										Name:     "bool",
										Type:     tftypes.Bool,
									},
									{
										Computed: true,
										Name:     "list",
										Optional: true,
										Type:     tftypes.List{ElementType: tftypes.String},
									},
									{
										Name:     "number",
										Optional: true,
										Type:     tftypes.Number,
									},
									{
										Name:     "string",
										Required: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
							TypeName: "list",
						},
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Computed: true,
										Name:     "bool",
										Type:     tftypes.Bool,
									},
									{
										Computed: true,
										Name:     "list",
										Optional: true,
										Type:     tftypes.List{ElementType: tftypes.String},
									},
									{
										Name:     "number",
										Optional: true,
										Type:     tftypes.Number,
									},
									{
										Name:     "string",
										Required: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
							TypeName: "set",
						},
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Computed: true,
										Name:     "bool",
										Type:     tftypes.Bool,
									},
									{
										Computed: true,
										Name:     "list",
										Optional: true,
										Type:     tftypes.List{ElementType: tftypes.String},
									},
									{
										Name:     "number",
										Optional: true,
										Type:     tftypes.Number,
									},
									{
										Name:     "string",
										Required: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
							TypeName: "single",
						},
					},
				},
			},
		},
		"markdown-description": {
			input: testschema.Schema{
				Version: 1,
				Attributes: map[string]fwschema.Attribute{
					"string": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				MarkdownDescription: "a test resource",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Description:     "a test resource",
					DescriptionKind: tfprotov6.StringKindMarkdown,
				},
			},
		},
		"plaintext-description": {
			input: testschema.Schema{
				Version: 1,
				Attributes: map[string]fwschema.Attribute{
					"string": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				Description: "a test resource",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Description:     "a test resource",
					DescriptionKind: tfprotov6.StringKindPlain,
				},
			},
		},
		"deprecated": {
			input: testschema.Schema{
				Version: 1,
				Attributes: map[string]fwschema.Attribute{
					"string": testschema.Attribute{
						Type:     types.StringType,
						Required: true,
					},
				},
				DeprecationMessage: "deprecated, use other_resource instead",
			},
			expected: &tfprotov6.Schema{
				Version: 1,
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "string",
							Type:     tftypes.String,
							Required: true,
						},
					},
					Deprecated: true,
				},
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := toproto6.Schema(context.Background(), tc.input)
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
			if err == nil && tc.expectedErr != "" {
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
