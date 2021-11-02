package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBlockTfprotov6(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		block       Block
		path        *tftypes.AttributePath
		expected    *tfprotov6.SchemaNestedBlock
		expectedErr string
	}

	tests := map[string]testCase{
		"nestingmode-invalid": {
			name: "test",
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
			},
			path:        tftypes.NewAttributePath(),
			expectedErr: "unrecognized nesting mode 0",
		},
		"nestingmode-list-attributes": {
			name: "test",
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				NestingMode: BlockNestingModeList,
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
			block: Block{
				Attributes: map[string]Attribute{
					"sub_attr": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"sub_block": {
						Attributes: map[string]Attribute{
							"sub_block_attr": {
								Type:     types.StringType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
				NestingMode: BlockNestingModeList,
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
			block: Block{
				Blocks: map[string]Block{
					"sub_block": {
						Attributes: map[string]Attribute{
							"sub_block_attr": {
								Type:     types.StringType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeList,
					},
				},
				NestingMode: BlockNestingModeList,
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
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				NestingMode: BlockNestingModeSet,
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
			block: Block{
				Attributes: map[string]Attribute{
					"sub_attr": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				Blocks: map[string]Block{
					"sub_block": {
						Attributes: map[string]Attribute{
							"sub_block_attr": {
								Type:     types.StringType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
				NestingMode: BlockNestingModeSet,
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
			block: Block{
				Blocks: map[string]Block{
					"sub_block": {
						Attributes: map[string]Attribute{
							"sub_block_attr": {
								Type:     types.StringType,
								Optional: true,
							},
						},
						NestingMode: BlockNestingModeSet,
					},
				},
				NestingMode: BlockNestingModeSet,
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
		"deprecationmessage": {
			name: "test",
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				DeprecationMessage: "deprecated, use something else instead",
				NestingMode:        BlockNestingModeList,
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
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				Description: "test description",
				NestingMode: BlockNestingModeList,
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
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				Description:         "test plain description",
				MarkdownDescription: "test markdown description",
				NestingMode:         BlockNestingModeList,
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
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				MarkdownDescription: "test description",
				NestingMode:         BlockNestingModeList,
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
		"maxitems": {
			name: "test",
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				MaxItems:    10,
				NestingMode: BlockNestingModeList,
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
		"minitems": {
			name: "test",
			block: Block{
				Attributes: map[string]Attribute{
					"sub_test": {
						Type:     types.StringType,
						Optional: true,
					},
				},
				MinItems:    10,
				NestingMode: BlockNestingModeList,
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
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.block.tfprotov6(context.Background(), tc.name, tc.path)
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

func TestBlockValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		req  ValidateAttributeRequest
		resp ValidateAttributeResponse
	}{
		"deprecation-message-known": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Block Deprecated",
						"Use something else instead.",
					),
				},
			},
		},
		"deprecation-message-null": {
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
								nil,
							),
						},
					),
					Schema: Schema{
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        BlockNestingModeList,
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
								tftypes.UnknownValue,
							),
						},
					),
					Schema: Schema{
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								DeprecationMessage: "Use something else instead.",
								NestingMode:        BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeWarningDiagnostic(
						tftypes.NewAttributePath().WithAttributeName("test"),
						"Block Deprecated",
						"Use something else instead.",
					),
				},
			},
		},
		"warnings": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: BlockNestingModeList,
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: BlockNestingModeList,
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
		"nested-attr-warnings": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testWarningAttributeValidator{},
											testWarningAttributeValidator{},
										},
									},
								},
								NestingMode: BlockNestingModeList,
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
		"nested-attr-errors": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
											testErrorAttributeValidator{},
										},
									},
								},
								NestingMode: BlockNestingModeList,
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
		"nested-attr-type-with-validate-error": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     testtypes.StringTypeWithValidateError{},
										Required: true,
									},
								},
								NestingMode: BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("nested_attr")),
				},
			},
		},
		"nested-attr-type-with-validate-warning": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     testtypes.StringTypeWithValidateWarning{},
										Required: true,
									},
								},
								NestingMode: BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{
				Diagnostics: diag.Diagnostics{
					testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("test").WithElementKeyInt(0).WithAttributeName("nested_attr")),
				},
			},
		},
		"list-no-validation": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: BlockNestingModeList,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"list-validation": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								},
								NestingMode: BlockNestingModeList,
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
		"set-no-validation": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
									},
								},
								NestingMode: BlockNestingModeSet,
							},
						},
					},
				},
			},
			resp: ValidateAttributeResponse{},
		},
		"set-validation": {
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
						Blocks: map[string]Block{
							"test": {
								Attributes: map[string]Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										Validators: []AttributeValidator{
											testErrorAttributeValidator{},
										},
									},
								},
								NestingMode: BlockNestingModeSet,
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
			block, err := tc.req.Config.Schema.blockAtPath(tc.req.AttributePath)

			if err != nil {
				t.Fatalf("Unexpected error getting %s", err)
			}

			block.validate(context.Background(), tc.req, &got)

			if diff := cmp.Diff(got, tc.resp); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}
