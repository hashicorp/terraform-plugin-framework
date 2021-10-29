package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAttributeTfprotov6(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		attr        Attribute
		path        *tftypes.AttributePath
		expected    interface{}
		expectedErr string
	}

	tests := map[string]testCase{
		"empty": {
			name:        "test",
			attr:        Attribute{},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Attributes, Blocks, or Type set",
		},
		"attributes": {
			name: "test",
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"sub_test": {
						Optional: true,
						Type:     types.StringType,
					},
				}, ListNestedAttributesOptions{}),
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name: "test",
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				Optional: true,
			},
		},
		"blocks": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Optional: true,
						Type:     types.StringType,
					},
				}, ListNestedBlocksOptions{}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"type": {
			name: "test",
			attr: Attribute{
				Optional: true,
				Type:     types.StringType,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "test",
				Optional: true,
				Type:     tftypes.String,
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.attr.tfprotov6(context.Background(), tc.name, tc.path)
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

func TestAttributeTfprotov6SchemaAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		attr        Attribute
		path        *tftypes.AttributePath
		expected    *tfprotov6.SchemaAttribute
		expectedErr string
	}

	tests := map[string]testCase{
		"deprecated": {
			name: "string",
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
		"attr-map": {
			name: "map",
			attr: Attribute{
				Type:     types.MapType{ElemType: types.StringType},
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "map",
				Type:     tftypes.Map{ElementType: tftypes.String},
				Optional: true,
			},
		},
		"attr-object": {
			name: "object",
			attr: Attribute{
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
		"attr-set": {
			name: "set",
			attr: Attribute{
				Type:     types.SetType{ElemType: types.NumberType},
				Optional: true,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaAttribute{
				Name:     "set",
				Type:     tftypes.Set{ElementType: tftypes.Number},
				Optional: true,
			},
		},
		// TODO: add tuple attribute when we support it
		"required": {
			name: "string",
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
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
			attr: Attribute{
				Attributes: SingleNestedAttributes(map[string]Attribute{
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
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, ListNestedAttributesOptions{}),
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
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, ListNestedAttributesOptions{
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
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, ListNestedAttributesOptions{
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
			attr: Attribute{
				Attributes: ListNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, ListNestedAttributesOptions{
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
			attr: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, SetNestedAttributesOptions{}),
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
			attr: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, SetNestedAttributesOptions{
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
			attr: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, SetNestedAttributesOptions{
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
			attr: Attribute{
				Attributes: SetNestedAttributes(map[string]Attribute{
					"string": {
						Type:     types.StringType,
						Optional: true,
					},
					"computed": {
						Type:      types.NumberType,
						Computed:  true,
						Sensitive: true,
					},
				}, SetNestedAttributesOptions{
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
			attr: Attribute{
				Type: types.StringType,
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"testing": {
						Type:     types.StringType,
						Optional: true,
					},
				}),
				Optional: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "cannot have both Attributes and Type set",
		},
		"attr-and-nested-attr-unset": {
			name: "whoops",
			attr: Attribute{
				Optional: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Attributes or Type set",
		},
		"attr-and-nested-attr-empty": {
			name: "whoops",
			attr: Attribute{
				Optional:   true,
				Attributes: SingleNestedAttributes(map[string]Attribute{}),
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Attributes or Type set",
		},
		"missing-required-optional-and-computed": {
			name: "whoops",
			attr: Attribute{
				Type: types.StringType,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "must have Required, Optional, or Computed set",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.attr.tfprotov6SchemaAttribute(context.Background(), tc.name, tc.path)
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

func TestAttributeTfprotov6SchemaNestedBlock(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		attr        Attribute
		path        *tftypes.AttributePath
		expected    *tfprotov6.SchemaNestedBlock
		expectedErr string
	}

	tests := map[string]testCase{
		"attributes": {
			name: "test",
			attr: Attribute{
				Attributes: SingleNestedAttributes(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}),
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "cannot have both Attributes and Blocks set",
		},
		"blocks-listnestedblocks": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"blocks-listnestedblocks-max": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{
					MaxItems: 10,
				}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				MaxItems: 10,
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"blocks-listnestedblocks-maxmin": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{
					MaxItems: 10,
					MinItems: 1,
				}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				MaxItems: 10,
				MinItems: 1,
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"blocks-listnestedblocks-min": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{
					MinItems: 10,
				}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				MinItems: 10,
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"blocks-setnestedblocks": {
			name: "test",
			attr: Attribute{
				Blocks: SetNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, SetNestedBlocksOptions{}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
				TypeName: "test",
			},
		},
		"blocks-setnestedblocks-max": {
			name: "test",
			attr: Attribute{
				Blocks: SetNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, SetNestedBlocksOptions{
					MaxItems: 10,
				}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				MaxItems: 10,
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
				TypeName: "test",
			},
		},
		"blocks-setnestedblocks-maxmin": {
			name: "test",
			attr: Attribute{
				Blocks: SetNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, SetNestedBlocksOptions{
					MaxItems: 10,
					MinItems: 1,
				}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				MaxItems: 10,
				MinItems: 1,
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
				TypeName: "test",
			},
		},
		"blocks-setnestedblocks-min": {
			name: "test",
			attr: Attribute{
				Blocks: SetNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, SetNestedBlocksOptions{
					MinItems: 10,
				}),
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
				},
				MinItems: 10,
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
				TypeName: "test",
			},
		},
		"computed": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				Computed: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "cannot set Block as Computed, mark all nested Attributes instead",
		},
		"deprecated": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				DeprecationMessage: "deprecated, use something else instead",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					Deprecated: true,
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"description-plain": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				Description: "test description",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					Description:     "test description",
					DescriptionKind: tfprotov6.StringKindPlain,
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"description-markdown": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				MarkdownDescription: "test description",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					Description:     "test description",
					DescriptionKind: tfprotov6.StringKindMarkdown,
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"description-both": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				Description:         "test plain description",
				MarkdownDescription: "test markdown description",
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_test",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					Description:     "test markdown description",
					DescriptionKind: tfprotov6.StringKindMarkdown,
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"optional": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				Optional: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "cannot set Block as Optional, mark all nested Attributes instead",
		},
		"required": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				Required: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "cannot set Block as Required, mark all nested Attributes instead",
		},
		"sensitive": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				Sensitive: true,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "cannot set Block as Sensitive, mark all nested Attributes instead",
		},
		"type": {
			name: "test",
			attr: Attribute{
				Blocks: ListNestedBlocks(map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				}, ListNestedBlocksOptions{}),
				Type: types.StringType,
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "cannot have both Blocks and Type set",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.attr.tfprotov6SchemaNestedBlock(context.Background(), tc.name, tc.path)
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

func TestAttributeModifyPlan(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req          ModifyAttributePlanRequest
		resp         ModifyAttributePlanResponse
		expectedResp ModifyAttributePlanResponse
	}{
		"config-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Configuration Read Error",
						"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
					),
				},
			},
		},
		"config-error-previous-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Configuration Read Error",
						"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
					),
				},
			},
		},
		"plan-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Plan Read Error",
						"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
					),
				},
			},
		},
		"plan-error-previous-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Plan Read Error",
						"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
					),
				},
			},
		},
		"state-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"State Read Error",
						"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
					),
				},
			},
		},
		"state-error-previous-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"State Read Error",
						"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
					),
				},
			},
		},
		"no-plan-modifiers": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
		},
		"attribute-plan": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testAttrPlanValueModifierOne{},
									testAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testAttrPlanValueModifierOne{},
									testAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testAttrPlanValueModifierOne{},
									testAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTATTRONE"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "MODIFIED_TWO"},
			},
		},
		"attribute-plan-previous-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testAttrPlanValueModifierOne{},
									testAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testAttrPlanValueModifierOne{},
									testAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testAttrPlanValueModifierOne{},
									testAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTATTRONE"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "MODIFIED_TWO"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
		},
		"requires-replacement": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan:   types.String{Value: "testvalue"},
				RequiresReplace: true,
			},
		},
		"requires-replacement-previous-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
				RequiresReplace: true,
			},
		},
		"requires-replacement-passthrough": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
									testAttrPlanValueModifierOne{},
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
									testAttrPlanValueModifierOne{},
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
									testAttrPlanValueModifierOne{},
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTATTRONE"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan:   types.String{Value: "TESTATTRTWO"},
				RequiresReplace: true,
			},
		},
		"requires-replacement-unset": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
									testRequiresReplaceFalseModifier{},
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
									testRequiresReplaceFalseModifier{},
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									RequiresReplace(),
									testRequiresReplaceFalseModifier{},
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
		},
		"warnings": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testWarningDiagModifier{},
									testWarningDiagModifier{},
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testWarningDiagModifier{},
									testWarningDiagModifier{},
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testWarningDiagModifier{},
									testWarningDiagModifier{},
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
				Diagnostics: diag.Diagnostics{
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
			},
		},
		"warnings-previous-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testWarningDiagModifier{},
									testWarningDiagModifier{},
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testWarningDiagModifier{},
									testWarningDiagModifier{},
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testWarningDiagModifier{},
									testWarningDiagModifier{},
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
			},
		},
		"error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testErrorDiagModifier{},
									testErrorDiagModifier{},
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testErrorDiagModifier{},
									testErrorDiagModifier{},
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testErrorDiagModifier{},
									testErrorDiagModifier{},
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
			},
		},
		"error-previous-error": {
			req: ModifyAttributePlanRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testErrorDiagModifier{},
									testErrorDiagModifier{},
								},
							},
						},
					},
				},
				Plan: Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testErrorDiagModifier{},
									testErrorDiagModifier{},
								},
							},
						},
					},
				},
				State: State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []AttributePlanModifier{
									testErrorDiagModifier{},
									testErrorDiagModifier{},
								},
							},
						},
					},
				},
			},
			resp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attribute, err := tc.req.Config.Schema.AttributeAtPath(tc.req.AttributePath)

			if err != nil {
				t.Fatalf("Unexpected error getting %s", err)
			}

			attribute.modifyPlan(context.Background(), tc.req, &tc.resp)

			if diff := cmp.Diff(tc.expectedResp, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestAttributeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req  ValidateAttributeRequest
		resp ValidateAttributeResponse
	}{
		"no-attributes-blocks-or-type": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Invalid Attribute Definition",
						"Attribute must define either Attributes, Blocks, or Type. This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"both-attributes-and-blocks": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SingleNestedAttributes(map[string]Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
									},
								}),
								Blocks: ListNestedBlocks(map[string]Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
									},
								}, ListNestedBlocksOptions{}),
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Invalid Attribute Definition",
						"Attribute cannot define both Attributes and Blocks. This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"both-attributes-and-type": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SingleNestedAttributes(map[string]Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
									},
								}),
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Invalid Attribute Definition",
						"Attribute cannot define both Attributes and Type. This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"both-blocks-and-type": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Blocks: ListNestedBlocks(map[string]Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
									},
								}, ListNestedBlocksOptions{}),
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Invalid Attribute Definition",
						"Attribute cannot define both Blocks and Type. This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"missing-required-optional-and-computed": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type: types.StringType,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Invalid Attribute Definition",
						"Attribute missing Required, Optional, or Computed definition. This is always a problem with the provider and should be reported to the provider developer.",
					),
				},
			},
		},
		"config-error": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Configuration Read Error",
						"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
					),
				},
			},
		},
		"no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"deprecation-message-known": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:               types.StringType,
								Optional:           true,
								DeprecationMessage: "Use something else instead.",
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Attribute Deprecated",
						"Use something else instead.",
					),
				},
			},
		},
		"deprecation-message-null": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:               types.StringType,
								Optional:           true,
								DeprecationMessage: "Use something else instead.",
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"deprecation-message-unknown": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:               types.StringType,
								Optional:           true,
								DeprecationMessage: "Use something else instead.",
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Attribute Deprecated",
						"Use something else instead.",
					),
				},
			},
		},
		"warnings": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testWarningAttributeValidator{},
									testWarningAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testWarningDiagnostic1,
					testWarningDiagnostic2,
				},
			},
		},
		"errors": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								Validators: []AttributeValidator{
									testErrorAttributeValidator{},
									testErrorAttributeValidator{},
								},
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
					testErrorDiagnostic2,
				},
			},
		},
		"type-with-validate-error": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     testtypes.StringTypeWithValidateError{},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("test")),
				},
			},
		},
		"type-with-validate-warning": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Type:     testtypes.StringTypeWithValidateWarning{},
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test")),
				},
			},
		},
		"nested-attr-list-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: ListNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}, ListNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-list-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: ListNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}, ListNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"nested-attr-map-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Map{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: MapNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}, MapNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-map-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Map{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Map{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								map[string]tftypes.Value{
									"testkey": tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: MapNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}, MapNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"nested-attr-set-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SetNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}, SetNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-set-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SetNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}, SetNestedAttributesOptions{}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"nested-attr-single-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SingleNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-attr-single-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
							},
						}, map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"nested_attr": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Attributes: SingleNestedAttributes(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}),
								Required: true,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"nested-block-list-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Blocks: ListNestedBlocks(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}, ListNestedBlocksOptions{}),
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-block-list-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Blocks: ListNestedBlocks(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}, ListNestedBlocksOptions{}),
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
		"nested-block-set-no-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Blocks: SetNestedBlocks(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								}, SetNestedBlocksOptions{}),
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"nested-block-set-validation": {
			req: ValidateAttributeRequest{
				AttributePath: tftypes.NewAttributePath().WithAttributeName("test"),
				Config: Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_attr": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_attr": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_attr": tftypes.NewValue(tftypes.String, "testvalue"),
										},
									),
								},
							),
						},
					),
					Schema: Schema{
						Attributes: map[string]Attribute{
							"test": {
								Blocks: SetNestedBlocks(map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								}, SetNestedBlocksOptions{}),
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testErrorDiagnostic1,
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got ValidateAttributeResponse
			attribute, err := tc.req.Config.Schema.AttributeAtPath(tc.req.AttributePath)

			if err != nil {
				t.Fatalf("Unexpected error getting %s", err)
			}

			attribute.validate(context.Background(), tc.req, &got)

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}
