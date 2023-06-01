// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBlock(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		block       fwschema.Block
		path        *tftypes.AttributePath
		expected    *tfprotov6.SchemaNestedBlock
		expectedErr string
	}

	tests := map[string]testCase{
		"nestingmode-invalid": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_test": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
				},
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "unrecognized nesting mode 0",
		},
		"nestingmode-list-attributes": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_test": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
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
		"nestingmode-list-attributes-and-blocks": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_attr": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"sub_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_block_attr": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_attr",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "sub_block_attr",
										Optional: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
							TypeName: "sub_block",
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"nestingmode-list-blocks": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Blocks: map[string]fwschema.Block{
						"sub_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_block_attr": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeList,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "sub_block_attr",
										Optional: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
							TypeName: "sub_block",
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
				TypeName: "test",
			},
		},
		"nestingmode-set-attributes": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_test": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeSet,
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
		"nestingmode-set-attributes-and-blocks": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_attr": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"sub_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_block_attr": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeSet,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_attr",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "sub_block_attr",
										Optional: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
							TypeName: "sub_block",
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
				TypeName: "test",
			},
		},
		"nestingmode-set-blocks": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Blocks: map[string]fwschema.Block{
						"sub_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_block_attr": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeSet,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "sub_block_attr",
										Optional: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
							TypeName: "sub_block",
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSet,
				TypeName: "test",
			},
		},
		"nestingmode-single-attributes": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_test": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeSingle,
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
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
				TypeName: "test",
			},
		},
		"nestingmode-single-attributes-and-blocks": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_attr": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"sub_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_block_attr": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeSingle,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "sub_attr",
							Optional: true,
							Type:     tftypes.String,
						},
					},
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "sub_block_attr",
										Optional: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
							TypeName: "sub_block",
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
				TypeName: "test",
			},
		},
		"nestingmode-single-blocks": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Blocks: map[string]fwschema.Block{
						"sub_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"sub_block_attr": testschema.Attribute{
										Type:     types.StringType,
										Optional: true,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSingle,
						},
					},
				},
				NestingMode: fwschema.BlockNestingModeSingle,
			},
			path: tftypes.NewAttributePath(),
			expected: &tfprotov6.SchemaNestedBlock{
				Block: &tfprotov6.SchemaBlock{
					BlockTypes: []*tfprotov6.SchemaNestedBlock{
						{
							Block: &tfprotov6.SchemaBlock{
								Attributes: []*tfprotov6.SchemaAttribute{
									{
										Name:     "sub_block_attr",
										Optional: true,
										Type:     tftypes.String,
									},
								},
							},
							Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
							TypeName: "sub_block",
						},
					},
				},
				Nesting:  tfprotov6.SchemaNestedBlockNestingModeSingle,
				TypeName: "test",
			},
		},
		"deprecationmessage": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_test": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
				},
				DeprecationMessage: "deprecated, use something else instead",
				NestingMode:        fwschema.BlockNestingModeList,
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
		"description": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_test": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
				},
				Description: "test description",
				NestingMode: fwschema.BlockNestingModeList,
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
		"description-and-markdowndescription": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_test": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
				},
				Description:         "test plain description",
				MarkdownDescription: "test markdown description",
				NestingMode:         fwschema.BlockNestingModeList,
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
		"markdowndescription": {
			name: "test",
			block: testschema.Block{
				NestedObject: testschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"sub_test": testschema.Attribute{
							Type:     types.StringType,
							Optional: true,
						},
					},
				},
				MarkdownDescription: "test description",
				NestingMode:         fwschema.BlockNestingModeList,
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
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := toproto6.Block(context.Background(), tc.name, tc.path, tc.block)
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
