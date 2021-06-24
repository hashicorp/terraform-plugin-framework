package proto6

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSchema(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       schema.Schema
		expected    *tfprotov6.Schema
		expectedErr string
	}

	tests := map[string]testCase{
		"empty-val": {
			input:       schema.Schema{},
			expectedErr: "must have at least one attribute in the schema",
		},
		"basic-attrs": {
			input: schema.Schema{
				Version: 1,
				Attributes: map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Required: true,
					},
					"number": {
						Type:     types.NumberType,
						Optional: true,
					},
					"bool": {
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
			input: schema.Schema{
				Version: 2,
				Attributes: map[string]schema.Attribute{
					"list": {
						Type:     types.ListType{ElemType: types.StringType},
						Required: true,
					},
					"object": {
						Type: types.ObjectType{AttrTypes: map[string]attr.Type{
							"string": types.StringType,
							"number": types.NumberType,
							"bool":   types.BoolType,
						}},
						Optional: true,
					},
					// TODO: add map support when it lands
					// TODO: add tuple support when it lands
					// TODO: add set support when it lands
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
							Name: "object",
							Type: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
								"string": tftypes.String,
								"number": tftypes.Number,
								"bool":   tftypes.Bool,
							}},
							Optional: true,
						},
					},
				},
			},
		},
		"nested-attrs": {
			input: schema.Schema{
				Version: 3,
				Attributes: map[string]schema.Attribute{
					"single": {
						Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
							"string": {
								Type:     types.StringType,
								Required: true,
							},
							"number": {
								Type:     types.NumberType,
								Optional: true,
							},
							"bool": {
								Type:     types.BoolType,
								Computed: true,
							},
							"list": {
								Type:     types.ListType{ElemType: types.StringType},
								Computed: true,
								Optional: true,
							},
						}),
						Required: true,
					},
					"list": {
						Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
							"string": {
								Type:     types.StringType,
								Required: true,
							},
							"number": {
								Type:     types.NumberType,
								Optional: true,
							},
							"bool": {
								Type:     types.BoolType,
								Computed: true,
							},
							"list": {
								Type:     types.ListType{ElemType: types.StringType},
								Computed: true,
								Optional: true,
							},
						}, schema.ListNestedAttributesOptions{}),
						Optional: true,
					},
					"set": {
						Attributes: schema.SetNestedAttributes(map[string]schema.Attribute{
							"string": {
								Type:     types.StringType,
								Required: true,
							},
							"number": {
								Type:     types.NumberType,
								Optional: true,
							},
							"bool": {
								Type:     types.BoolType,
								Computed: true,
							},
							"list": {
								Type:     types.ListType{ElemType: types.StringType},
								Computed: true,
								Optional: true,
							},
						}, schema.SetNestedAttributesOptions{}),
						Computed: true,
					},
					"map": {
						Attributes: schema.MapNestedAttributes(map[string]schema.Attribute{
							"string": {
								Type:     types.StringType,
								Required: true,
							},
							"number": {
								Type:     types.NumberType,
								Optional: true,
							},
							"bool": {
								Type:     types.BoolType,
								Computed: true,
							},
							"list": {
								Type:     types.ListType{ElemType: types.StringType},
								Computed: true,
								Optional: true,
							},
						}, schema.MapNestedAttributesOptions{}),
						Optional: true,
						Computed: true,
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
		"markdown-description": {
			input: schema.Schema{
				Version: 1,
				Attributes: map[string]schema.Attribute{
					"string": {
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
			input: schema.Schema{
				Version: 1,
				Attributes: map[string]schema.Attribute{
					"string": {
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
			input: schema.Schema{
				Version: 1,
				Attributes: map[string]schema.Attribute{
					"string": {
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

			got, err := Schema(context.Background(), tc.input)
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

func TestAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		attr        schema.Attribute
		path        *tftypes.AttributePath
		expected    *tfprotov6.SchemaAttribute
		expectedErr string
	}

	tests := map[string]testCase{
		"deprecated": {
			name: "string",
			attr: schema.Attribute{
				Type:               types.StringType,
				Optional:           true,
				DeprecationMessage: "deprecated, use new_string instead",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:       "string",
				Type:       tftypes.String,
				Optional:   true,
				Deprecated: true,
			},
		},
		"description-plain": {
			name: "string",
			attr: schema.Attribute{
				Type:        types.StringType,
				Optional:    true,
				Description: "A string attribute",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:            "string",
				Type:            tftypes.String,
				Optional:        true,
				Description:     "A string attribute",
				DescriptionKind: tfprotov6.StringKindPlain,
			},
		},
		"description-markdown": {
			name: "string",
			attr: schema.Attribute{
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "A string attribute",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:            "string",
				Type:            tftypes.String,
				Optional:        true,
				Description:     "A string attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
		},
		"description-both": {
			name: "string",
			attr: schema.Attribute{
				Type:                types.StringType,
				Optional:            true,
				Description:         "A string attribute",
				MarkdownDescription: "A string attribute (markdown)",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:            "string",
				Type:            tftypes.String,
				Optional:        true,
				Description:     "A string attribute (markdown)",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
		},
		"attr-string": {
			name: "string",
			attr: schema.Attribute{
				Type:     types.StringType,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Optional: true,
			},
		},
		"attr-bool": {
			name: "bool",
			attr: schema.Attribute{
				Type:     types.BoolType,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "bool",
				Type:     tftypes.Bool,
				Optional: true,
			},
		},
		"attr-number": {
			name: "number",
			attr: schema.Attribute{
				Type:     types.NumberType,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "number",
				Type:     tftypes.Number,
				Optional: true,
			},
		},
		"attr-list": {
			name: "list",
			attr: schema.Attribute{
				Type:     types.ListType{ElemType: types.NumberType},
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list",
				Type:     tftypes.List{ElementType: tftypes.Number},
				Optional: true,
			},
		},
		"attr-object": {
			name: "object",
			attr: schema.Attribute{
				Type: types.ObjectType{AttrTypes: map[string]attr.Type{
					"foo": types.StringType,
					"bar": types.NumberType,
					"baz": types.BoolType,
				}},
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name: "object",
				Type: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
					"baz": tftypes.Bool,
				}},
				Optional: true,
			},
		},
		// TODO: add map attribute when we support it
		// TODO: add set attribute when we support it
		// TODO: add tuple attribute when we support it
		"required": {
			name: "string",
			attr: schema.Attribute{
				Type:     types.StringType,
				Required: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Required: true,
			},
		},
		"optional": {
			name: "string",
			attr: schema.Attribute{
				Type:     types.StringType,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Optional: true,
			},
		},
		"computed": {
			name: "string",
			attr: schema.Attribute{
				Type:     types.StringType,
				Computed: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Computed: true,
			},
		},
		"optional-computed": {
			name: "string",
			attr: schema.Attribute{
				Type:     types.StringType,
				Computed: true,
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "string",
				Type:     tftypes.String,
				Computed: true,
				Optional: true,
			},
		},
		"sensitive": {
			name: "string",
			attr: schema.Attribute{
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:      "string",
				Type:      tftypes.String,
				Optional:  true,
				Sensitive: true,
			},
		},
		"nested-attr-single": {
			name: "single_nested",
			attr: schema.Attribute{
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "single_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSingle,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
			},
		},
		"nested-attr-list": {
			name: "list_nested",
			attr: schema.Attribute{
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, schema.ListNestedAttributesOptions{}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
			},
		},
		"nested-attr-list-min": {
			name: "list_nested",
			attr: schema.Attribute{
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, schema.ListNestedAttributesOptions{
					MinItems: 1,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MinItems: 1,
				},
			},
		},
		"nested-attr-list-max": {
			name: "list_nested",
			attr: schema.Attribute{
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, schema.ListNestedAttributesOptions{
					MaxItems: 1,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MaxItems: 1,
				},
			},
		},
		"nested-attr-list-minmax": {
			name: "list_nested",
			attr: schema.Attribute{
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, schema.ListNestedAttributesOptions{
					MinItems: 1,
					MaxItems: 10,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "list_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MinItems: 1,
					MaxItems: 10,
				},
			},
		},
		"nested-attr-set": {
			name: "set_nested",
			attr: schema.Attribute{
				Attributes: schema.SetNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, schema.SetNestedAttributesOptions{}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
			},
		},
		"nested-attr-set-min": {
			name: "set_nested",
			attr: schema.Attribute{
				Attributes: schema.SetNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, schema.SetNestedAttributesOptions{
					MinItems: 1,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MinItems: 1,
				},
			},
		},
		"nested-attr-set-max": {
			name: "set_nested",
			attr: schema.Attribute{
				Attributes: schema.SetNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, schema.SetNestedAttributesOptions{
					MaxItems: 1,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MaxItems: 1,
				},
			},
		},
		"nested-attr-set-minmax": {
			name: "set_nested",
			attr: schema.Attribute{
				Attributes: schema.SetNestedAttributes(map[string]schema.Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, schema.SetNestedAttributesOptions{
					MinItems: 1,
					MaxItems: 10,
				}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set_nested",
				Optional: true,
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSet,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:      "computed",
							Computed:  true,
							Sensitive: true,
							Type:      tftypes.Number,
						},
						{
							Name:     "string",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					MinItems: 1,
					MaxItems: 10,
				},
			},
		},
		"attr-and-nested-attr-set": {
			name: "whoops",
			attr: schema.Attribute{
				Type: types.StringType,
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
					"testing": {
						Type:     types.StringType,
						Optional: true,
					},
				}),
				Optional: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "can't have both Attributes and Type set",
		},
		"attr-and-nested-attr-unset": {
			name: "whoops",
			attr: schema.Attribute{
				Optional: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Attributes or Type set",
		},
		"attr-and-nested-attr-empty": {
			name: "whoops",
			attr: schema.Attribute{
				Optional:   true,
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{}),
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Attributes or Type set",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := Attribute(context.Background(), tc.name, tc.attr, tc.path)
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
