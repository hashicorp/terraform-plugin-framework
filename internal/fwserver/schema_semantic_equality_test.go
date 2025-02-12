// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSchemaSemanticEquality(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwserver.SchemaSemanticEqualityRequest
		expected *fwserver.SchemaSemanticEqualityResponse
	}{
		"Attribute-Valuable": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
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
							"test": tftypes.NewValue(tftypes.String, "prior"),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
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
							"test": tftypes.NewValue(tftypes.String, "new"),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
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
							"test": tftypes.NewValue(tftypes.String, "new"),
						},
					),
				},
			},
		},
		"Attribute-ValuableWithSemanticEquals-true": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
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
							"test": tftypes.NewValue(tftypes.String, "prior"),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
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
							"test": tftypes.NewValue(tftypes.String, "new"),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
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
							"test": tftypes.NewValue(tftypes.String, "prior"),
						},
					),
				},
			},
		},
		"Attribute-ValuableWithSemanticEquals-false": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
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
							"test": tftypes.NewValue(tftypes.String, "prior"),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
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
							"test": tftypes.NewValue(tftypes.String, "new"),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
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
							"test": tftypes.NewValue(tftypes.String, "new"),
						},
					),
				},
			},
		},
		"Attribute-ValuableWithSemanticEquals-diagnostics": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
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
							"test": tftypes.NewValue(tftypes.String, "prior"),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
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
							"test": tftypes.NewValue(tftypes.String, "new"),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Optional: true,
								Type: testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
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
							"test": tftypes.NewValue(tftypes.String, "new"),
						},
					),
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"Block-List-Valuable": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
		},
		"Block-List-ValuableWithSemanticEquals-true": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
			},
		},
		"Block-List-ValuableWithSemanticEquals-false": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
		},
		"Block-List-ValuableWithSemanticEquals-diagnostics": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
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
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"Block-Set-Valuable": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
		},
		"Block-Set-ValuableWithSemanticEquals-true": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
			},
		},
		"Block-Set-ValuableWithSemanticEquals-false": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
		},
		"Block-Set-ValuableWithSemanticEquals-diagnostics": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "prior"),
										},
									),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
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
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"test": tftypes.String,
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
											"test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"test": tftypes.NewValue(tftypes.String, "new"),
										},
									),
								},
							),
						},
					),
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		"Block-Single-Valuable": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type:     types.StringType,
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "prior"),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type:     types.StringType,
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "new"),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type:     types.StringType,
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "new"),
								},
							),
						},
					),
				},
			},
		},
		"Block-Single-ValuableWithSemanticEquals-true": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "prior"),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "new"),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: true,
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "prior"),
								},
							),
						},
					),
				},
			},
		},
		"Block-Single-ValuableWithSemanticEquals-false": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "prior"),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "new"),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "new"),
								},
							),
						},
					),
				},
			},
		},
		"Block-Single-ValuableWithSemanticEquals-diagnostics": {
			request: fwserver.SchemaSemanticEqualityRequest{
				PriorData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionState,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "prior"),
								},
							),
						},
					),
				},
				ProposedNewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "new"),
								},
							),
						},
					),
				},
			},
			expected: &fwserver.SchemaSemanticEqualityResponse{
				NewData: fwschemadata.Data{
					Description: fwschemadata.DataDescriptionPlan,
					Schema: testschema.Schema{
						Blocks: map[string]fwschema.Block{
							"test": testschema.Block{
								NestedObject: testschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"test": testschema.Attribute{
											Optional: true,
											Type: testtypes.StringTypeWithSemanticEquals{
												SemanticEquals: false,
												SemanticEqualsDiagnostics: diag.Diagnostics{
													diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
													diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
												},
											},
										},
									},
								},
								NestingMode: fwschema.BlockNestingModeSingle,
							},
						},
					},
					TerraformValue: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
							},
						},
						map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"test": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"test": tftypes.NewValue(tftypes.String, "new"),
								},
							),
						},
					),
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwserver.SchemaSemanticEqualityResponse{
				NewData: testCase.request.ProposedNewData,
			}

			fwserver.SchemaSemanticEquality(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
