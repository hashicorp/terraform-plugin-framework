// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/toproto6"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// DynamicValueMust returns a *tfprotov6.DynamicValue or panics.
func DynamicValueMust(value tftypes.Value) *tfprotov6.DynamicValue {
	dynamicValue, err := tfprotov6.NewDynamicValue(value.Type(), value)

	if err != nil {
		panic(err)
	}

	return &dynamicValue
}

func TestDynamicValue(t *testing.T) {
	t.Parallel()

	// NOTE: These test cases are not intended to be exhaustive for potential
	// roundtrips of tftypes.Value <=> *tfprotov6.DynamicValue. Rather, each
	// test case should only be present for this package's modifications of the
	// data, similar edge cases to those modifications, or regressions.
	testCases := map[string]struct {
		fw            *fwschemadata.Data
		expected      *tfprotov6.DynamicValue
		expectedDiags diag.Diagnostics
	}{
		"nil": {
			fw:       nil,
			expected: nil,
		},
		"NewDynamicValue-error": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Optional: true,
							Type:     types.BoolType, // intentional for testing error
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "test-value"),
					},
				),
			},
			expected: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Configuration",
					"An unexpected error was encountered when converting the configuration to the protocol type. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Unable to create DynamicValue: AttributeName(\"test\"): unexpected value type string, tftypes.Bool values must be of type bool",
				),
			},
		},
		"attribute-value": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "test-value"),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.String, "test-value"),
				},
			)),
		},
		"block-list-empty": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
					Blocks: map[string]fwschema.Block{
						"test_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"test_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_attribute": tftypes.String,
							"test_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
						"test_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
							},
							nil, // should be converted to empty value
						),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
					"test_block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{},
					),
				},
			)),
		},
		"block-list-value": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
					Blocks: map[string]fwschema.Block{
						"test_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"test_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_attribute": tftypes.String,
							"test_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
						"test_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_block_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
									},
								),
							},
						),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
					"test_block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
								},
							),
						},
					),
				},
			)),
		},
		"block-nested-block-list-empty": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
					Blocks: map[string]fwschema.Block{
						"test_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"test_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
								Blocks: map[string]fwschema.Block{
									"test_nested_block": testschema.Block{
										NestedObject: testschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"test_nested_block_attribute": testschema.Attribute{
													Optional: true,
													Type:     types.StringType,
												},
											},
										},
										NestingMode: fwschema.BlockNestingModeList,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_attribute": tftypes.String,
							"test_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.List{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
						"test_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.List{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_block_attribute": tftypes.String,
											"test_nested_block": tftypes.List{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
											},
										},
									},
									map[string]tftypes.Value{
										"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
										"test_nested_block": tftypes.NewValue(
											tftypes.List{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
											},
											nil, // should be converted to empty value
										),
									},
								),
							},
						),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
									"test_nested_block": tftypes.List{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_nested_block_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
					"test_block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
									"test_nested_block": tftypes.List{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_nested_block_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.List{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
								map[string]tftypes.Value{
									"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
									"test_nested_block": tftypes.NewValue(
										tftypes.List{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
										[]tftypes.Value{},
									),
								},
							),
						},
					),
				},
			)),
		},
		"block-nested-block-list-value": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
					Blocks: map[string]fwschema.Block{
						"test_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"test_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
								Blocks: map[string]fwschema.Block{
									"test_nested_block": testschema.Block{
										NestedObject: testschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"test_nested_block_attribute": testschema.Attribute{
													Optional: true,
													Type:     types.StringType,
												},
											},
										},
										NestingMode: fwschema.BlockNestingModeList,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeList,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_attribute": tftypes.String,
							"test_block": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.List{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
						"test_block": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.List{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_block_attribute": tftypes.String,
											"test_nested_block": tftypes.List{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
											},
										},
									},
									map[string]tftypes.Value{
										"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
										"test_nested_block": tftypes.NewValue(
											tftypes.List{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
											},
											[]tftypes.Value{
												tftypes.NewValue(
													tftypes.Object{
														AttributeTypes: map[string]tftypes.Type{
															"test_nested_block_attribute": tftypes.String,
														},
													},
													map[string]tftypes.Value{
														"test_nested_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
													},
												),
											},
										),
									},
								),
							},
						),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
									"test_nested_block": tftypes.List{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_nested_block_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
					"test_block": tftypes.NewValue(
						tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
									"test_nested_block": tftypes.List{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_nested_block_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.List{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
								map[string]tftypes.Value{
									"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
									"test_nested_block": tftypes.NewValue(
										tftypes.List{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
										[]tftypes.Value{
											tftypes.NewValue(
												tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
												map[string]tftypes.Value{
													"test_nested_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
												},
											),
										},
									),
								},
							),
						},
					),
				},
			)),
		},
		"block-set-empty": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
					Blocks: map[string]fwschema.Block{
						"test_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"test_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_attribute": tftypes.String,
							"test_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
						"test_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
							},
							nil, // should be converted to empty value
						),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
					"test_block": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{},
					),
				},
			)),
		},
		"block-set-value": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
					Blocks: map[string]fwschema.Block{
						"test_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"test_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_attribute": tftypes.String,
							"test_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
						"test_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_block_attribute": tftypes.String,
										},
									},
									map[string]tftypes.Value{
										"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
									},
								),
							},
						),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
					"test_block": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
								},
							),
						},
					),
				},
			)),
		},
		"block-nested-block-set-empty": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
					Blocks: map[string]fwschema.Block{
						"test_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"test_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
								Blocks: map[string]fwschema.Block{
									"test_nested_block": testschema.Block{
										NestedObject: testschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"test_nested_block_attribute": testschema.Attribute{
													Optional: true,
													Type:     types.StringType,
												},
											},
										},
										NestingMode: fwschema.BlockNestingModeSet,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_attribute": tftypes.String,
							"test_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.Set{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
						"test_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.Set{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_block_attribute": tftypes.String,
											"test_nested_block": tftypes.Set{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
											},
										},
									},
									map[string]tftypes.Value{
										"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
										"test_nested_block": tftypes.NewValue(
											tftypes.Set{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
											},
											nil, // should be converted to empty value
										),
									},
								),
							},
						),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
									"test_nested_block": tftypes.Set{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_nested_block_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
					"test_block": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
									"test_nested_block": tftypes.Set{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_nested_block_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.Set{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
								map[string]tftypes.Value{
									"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
									"test_nested_block": tftypes.NewValue(
										tftypes.Set{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
										[]tftypes.Value{},
									),
								},
							),
						},
					),
				},
			)),
		},
		"block-nested-block-set-value": {
			fw: &fwschemadata.Data{
				Description: fwschemadata.DataDescriptionConfiguration,
				Schema: testschema.Schema{
					Attributes: map[string]fwschema.Attribute{
						"test_attribute": testschema.Attribute{
							Optional: true,
							Type:     types.StringType,
						},
					},
					Blocks: map[string]fwschema.Block{
						"test_block": testschema.Block{
							NestedObject: testschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"test_block_attribute": testschema.Attribute{
										Optional: true,
										Type:     types.StringType,
									},
								},
								Blocks: map[string]fwschema.Block{
									"test_nested_block": testschema.Block{
										NestedObject: testschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"test_nested_block_attribute": testschema.Attribute{
													Optional: true,
													Type:     types.StringType,
												},
											},
										},
										NestingMode: fwschema.BlockNestingModeSet,
									},
								},
							},
							NestingMode: fwschema.BlockNestingModeSet,
						},
					},
				},
				TerraformValue: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test_attribute": tftypes.String,
							"test_block": tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.Set{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
						"test_block": tftypes.NewValue(
							tftypes.Set{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.Set{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test_block_attribute": tftypes.String,
											"test_nested_block": tftypes.Set{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
											},
										},
									},
									map[string]tftypes.Value{
										"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
										"test_nested_block": tftypes.NewValue(
											tftypes.Set{
												ElementType: tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
											},
											[]tftypes.Value{
												tftypes.NewValue(
													tftypes.Object{
														AttributeTypes: map[string]tftypes.Type{
															"test_nested_block_attribute": tftypes.String,
														},
													},
													map[string]tftypes.Value{
														"test_nested_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
													},
												),
											},
										),
									},
								),
							},
						),
					},
				),
			},
			expected: DynamicValueMust(tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test_attribute": tftypes.String,
						"test_block": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
									"test_nested_block": tftypes.Set{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_nested_block_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
					},
				},
				map[string]tftypes.Value{
					"test_attribute": tftypes.NewValue(tftypes.String, "test-value"),
					"test_block": tftypes.NewValue(
						tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"test_block_attribute": tftypes.String,
									"test_nested_block": tftypes.Set{
										ElementType: tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test_nested_block_attribute": tftypes.String,
											},
										},
									},
								},
							},
						},
						[]tftypes.Value{
							tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test_block_attribute": tftypes.String,
										"test_nested_block": tftypes.Set{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
									},
								},
								map[string]tftypes.Value{
									"test_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
									"test_nested_block": tftypes.NewValue(
										tftypes.Set{
											ElementType: tftypes.Object{
												AttributeTypes: map[string]tftypes.Type{
													"test_nested_block_attribute": tftypes.String,
												},
											},
										},
										[]tftypes.Value{
											tftypes.NewValue(
												tftypes.Object{
													AttributeTypes: map[string]tftypes.Type{
														"test_nested_block_attribute": tftypes.String,
													},
												},
												map[string]tftypes.Value{
													"test_nested_block_attribute": tftypes.NewValue(tftypes.String, "test-value"),
												},
											),
										},
									),
								},
							),
						},
					),
				},
			)),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := toproto6.DynamicValue(context.Background(), testCase.fw)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
