package fwplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestMatchElementStateForUnknownModifierPlanModify(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request     fwplanmodifier.MatchElementStateForUnknownRequest
		expressions path.Expressions
		expected    *fwplanmodifier.MatchElementStateForUnknownResponse
	}{
		"path-root": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:      path.Root("test"),
				PlanValue: types.StringUnknown(),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownOutsideListOrSetDiag(
						path.Root("test"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"path-map": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:      path.Root("test").AtMapKey("testkey").AtName("nested_test"),
				PlanValue: types.StringUnknown(),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownOutsideListOrSetDiag(
						path.Root("test").AtMapKey("testkey").AtName("nested_test"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"path-object": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:      path.Root("test").AtName("nested_test"),
				PlanValue: types.StringUnknown(),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownOutsideListOrSetDiag(
						path.Root("test").AtName("nested_test"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"path-list-no-expressions": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:      path.Root("test").AtListIndex(0).AtName("nested_test"),
				PlanValue: types.StringUnknown(),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownMissingExpressionsDiag(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"path-set-no-expressions": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path: path.Root("test").AtSetValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_test": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_test": types.StringType,
								},
								map[string]attr.Value{
									"nested_test": types.StringUnknown(),
								},
							),
						},
					),
				).AtName("nested_test"),
				PlanValue: types.StringUnknown(),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownMissingExpressionsDiag(
						path.Root("test").AtSetValue(
							types.SetValueMust(
								types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"nested_test": types.StringType,
									},
								},
								[]attr.Value{
									types.ObjectValueMust(
										map[string]attr.Type{
											"nested_test": types.StringType,
										},
										map[string]attr.Value{
											"nested_test": types.StringUnknown(),
										},
									),
								},
							),
						).AtName("nested_test"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"expressions-root": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:      types.StringUnknown(),
			},
			expressions: path.Expressions{
				// not valid for multiple elements
				path.MatchRoot("test").AtListIndex(0).AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownRootExpressionDiag(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
						path.MatchRoot("test").AtListIndex(0).AtName("nested_other"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"expressions-invalid-step-length": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:      types.StringUnknown(),
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other").AtName("oops"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownInvalidExpressionDiag(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
						path.MatchRelative().AtParent().AtName("nested_other").AtName("oops"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"expressions-invalid-step-0": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:      types.StringUnknown(),
			},
			expressions: path.Expressions{
				path.MatchRelative().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownInvalidExpressionDiag(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
						path.MatchRelative().AtName("nested_other"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"expressions-invalid-step-1": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:      types.StringUnknown(),
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtParent(),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownInvalidExpressionDiag(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
						path.MatchRelative().AtParent().AtParent(),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"expressions-self-reference": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_test": tftypes.String,
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
											"nested_test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_test"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_test"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					fwplanmodifier.MatchElementStateForUnknownInvalidExpressionDiag(
						path.Root("test").AtListIndex(0).AtName("nested_test"),
						path.MatchRoot("test").AtListIndex(0).AtName("nested_test").AtParent().AtName("nested_test"),
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"expressions-no-matched-path": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											// intentionally missing nested_other
											"nested_test": tftypes.String,
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
											// intentionally missing nested_other
											"nested_test": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												// intentionally missing nested_other
												"nested_test": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_test": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										// intentionally missing nested_other
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid Path Expression for Schema",
						"The Terraform Provider unexpectedly provided a path expression that does not match the current schema. "+
							"This can happen if the path expression does not correctly follow the schema in structure or types. "+
							"Please report this to the provider developers.\n\n"+
							"Path Expression: test[0].nested_test.<.nested_other",
					),
				},
				PlanValue: types.StringUnknown(),
			},
		},
		"unknown-identifying-value": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue:    types.StringNull(),
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_test":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, "teststatevalue"),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringUnknown(),
			},
		},
		"known-plan-value": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue:    types.StringNull(),
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, "knownplanvalue"),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringValue("knownplanvalue"),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, "teststatevalue"),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringValue("knownplanvalue"),
			},
		},
		"unknown-config-value": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue:    types.StringUnknown(),
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Optional: true,
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, "teststatevalue"),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringUnknown(),
			},
		},
		"null-matching-prior-state-value": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue:    types.StringNull(),
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, nil),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringUnknown(),
			},
		},
		"null-state": { // resource creation
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue:    types.StringNull(),
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
							},
						},
						nil,
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringUnknown(),
			},
		},
		"list-matching-prior-state-value": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue:    types.StringNull(),
				Path:           path.Root("test").AtListIndex(0).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, "teststatevalue"),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringValue("teststatevalue"),
			},
		},
		"list-rearranged": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue:    types.StringNull(),
				Path:           path.Root("test").AtListIndex(0).AtName("nested_computed1"),
				PathExpression: path.MatchRoot("test").AtListIndex(0).AtName("nested_computed1"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_computed1":   tftypes.String,
											"nested_computed2":   tftypes.String,
											"nested_configured1": tftypes.String,
											"nested_configured2": tftypes.String,
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
											"nested_computed1":   tftypes.String,
											"nested_computed2":   tftypes.String,
											"nested_configured1": tftypes.String,
											"nested_configured2": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed1":   tftypes.String,
												"nested_computed2":   tftypes.String,
												"nested_configured1": tftypes.String,
												"nested_configured2": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_configured1": tftypes.NewValue(tftypes.String, "element2-configured1"),
											"nested_configured2": tftypes.NewValue(tftypes.String, "element2-configured2"),
											"nested_computed1":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_computed2":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed1":   tftypes.String,
												"nested_computed2":   tftypes.String,
												"nested_configured1": tftypes.String,
												"nested_configured2": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_configured1": tftypes.NewValue(tftypes.String, "element1-configured1"),
											"nested_configured2": tftypes.NewValue(tftypes.String, "element1-configured2"),
											"nested_computed1":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_computed2":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_computed1": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_configured1"),
														path.MatchRelative().AtParent().AtName("nested_configured2"),
													},
												},
											},
										},
										"nested_computed2": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
										},
										"nested_configured1": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_configured2": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.List{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_computed1":   tftypes.String,
											"nested_computed2":   tftypes.String,
											"nested_configured1": tftypes.String,
											"nested_configured2": tftypes.String,
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
											"nested_computed1":   tftypes.String,
											"nested_computed2":   tftypes.String,
											"nested_configured1": tftypes.String,
											"nested_configured2": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed1":   tftypes.String,
												"nested_computed2":   tftypes.String,
												"nested_configured1": tftypes.String,
												"nested_configured2": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_configured1": tftypes.NewValue(tftypes.String, "element1-configured1"),
											"nested_configured2": tftypes.NewValue(tftypes.String, "element1-configured2"),
											"nested_computed1":   tftypes.NewValue(tftypes.String, "element1-computed1"),
											"nested_computed2":   tftypes.NewValue(tftypes.String, "element1-computed2"),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed1":   tftypes.String,
												"nested_computed2":   tftypes.String,
												"nested_configured1": tftypes.String,
												"nested_configured2": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_configured1": tftypes.NewValue(tftypes.String, "element2-configured1"),
											"nested_configured2": tftypes.NewValue(tftypes.String, "element2-configured2"),
											"nested_computed1":   tftypes.NewValue(tftypes.String, "element2-computed1"),
											"nested_computed2":   tftypes.NewValue(tftypes.String, "element2-computed2"),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_computed1": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_configured1"),
														path.MatchRelative().AtParent().AtName("nested_configured2"),
													},
												},
											},
										},
										"nested_computed2": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
										},
										"nested_configured1": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_configured2": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeList,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_configured1"),
				path.MatchRelative().AtParent().AtName("nested_configured2"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringValue("element2-computed1"),
			},
		},
		"set-matching-prior-state-value": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue: types.StringNull(),
				Path: path.Root("test").AtSetValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_other": types.StringType,
							"nested_test":  types.StringType,
						},
						map[string]attr.Value{
							"nested_other": types.StringValue("otherstatevalue"),
							"nested_test":  types.StringUnknown(),
						},
					),
				).AtName("nested_test"),
				PathExpression: path.MatchRoot("test").AtSetValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_other": types.StringType,
							"nested_test":  types.StringType,
						},
						map[string]attr.Value{
							"nested_other": types.StringValue("otherstatevalue"),
							"nested_test":  types.StringUnknown(),
						},
					),
				).AtName("nested_test"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
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
											"nested_other": tftypes.String,
											"nested_test":  tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_other": tftypes.String,
												"nested_test":  tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_other": tftypes.NewValue(tftypes.String, "otherstatevalue"),
											"nested_test":  tftypes.NewValue(tftypes.String, "teststatevalue"),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_other": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_test": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_other"),
													},
												},
											},
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_other"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringValue("teststatevalue"),
			},
		},
		"set-rearranged": {
			request: fwplanmodifier.MatchElementStateForUnknownRequest{
				ConfigValue: types.StringNull(),
				Path: path.Root("test").AtSetValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_computed1":   types.StringType,
							"nested_computed2":   types.StringType,
							"nested_configured1": types.StringType,
							"nested_configured2": types.StringType,
						},
						map[string]attr.Value{
							"nested_computed1":   types.StringUnknown(),
							"nested_computed2":   types.StringUnknown(),
							"nested_configured1": types.StringValue("element2-configured1"),
							"nested_configured2": types.StringValue("element2-configured2"),
						},
					),
				).AtName("nested_computed1"),
				PathExpression: path.MatchRoot("test").AtSetValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"nested_computed1":   types.StringType,
							"nested_computed2":   types.StringType,
							"nested_configured1": types.StringType,
							"nested_configured2": types.StringType,
						},
						map[string]attr.Value{
							"nested_computed1":   types.StringUnknown(),
							"nested_computed2":   types.StringUnknown(),
							"nested_configured1": types.StringValue("element2-configured1"),
							"nested_configured2": types.StringValue("element2-configured2"),
						},
					),
				).AtName("nested_computed1"),
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_computed1":   tftypes.String,
											"nested_computed2":   tftypes.String,
											"nested_configured1": tftypes.String,
											"nested_configured2": tftypes.String,
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
											"nested_computed1":   tftypes.String,
											"nested_computed2":   tftypes.String,
											"nested_configured1": tftypes.String,
											"nested_configured2": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed1":   tftypes.String,
												"nested_computed2":   tftypes.String,
												"nested_configured1": tftypes.String,
												"nested_configured2": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_configured1": tftypes.NewValue(tftypes.String, "element2-configured1"),
											"nested_configured2": tftypes.NewValue(tftypes.String, "element2-configured2"),
											"nested_computed1":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_computed2":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed1":   tftypes.String,
												"nested_computed2":   tftypes.String,
												"nested_configured1": tftypes.String,
												"nested_configured2": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_configured1": tftypes.NewValue(tftypes.String, "element1-configured1"),
											"nested_configured2": tftypes.NewValue(tftypes.String, "element1-configured2"),
											"nested_computed1":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_computed2":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_computed1": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_configured1"),
														path.MatchRelative().AtParent().AtName("nested_configured2"),
													},
												},
											},
										},
										"nested_computed2": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
										},
										"nested_configured1": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_configured2": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
							},
						},
					},
				},
				PlanValue: types.StringUnknown(),
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_computed1":   tftypes.String,
											"nested_computed2":   tftypes.String,
											"nested_configured1": tftypes.String,
											"nested_configured2": tftypes.String,
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
											"nested_computed1":   tftypes.String,
											"nested_computed2":   tftypes.String,
											"nested_configured1": tftypes.String,
											"nested_configured2": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed1":   tftypes.String,
												"nested_computed2":   tftypes.String,
												"nested_configured1": tftypes.String,
												"nested_configured2": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_configured1": tftypes.NewValue(tftypes.String, "element1-configured1"),
											"nested_configured2": tftypes.NewValue(tftypes.String, "element1-configured2"),
											"nested_computed1":   tftypes.NewValue(tftypes.String, "element1-computed1"),
											"nested_computed2":   tftypes.NewValue(tftypes.String, "element1-computed2"),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed1":   tftypes.String,
												"nested_computed2":   tftypes.String,
												"nested_configured1": tftypes.String,
												"nested_configured2": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_configured1": tftypes.NewValue(tftypes.String, "element2-configured1"),
											"nested_configured2": tftypes.NewValue(tftypes.String, "element2-configured2"),
											"nested_computed1":   tftypes.NewValue(tftypes.String, "element2-computed1"),
											"nested_computed2":   tftypes.NewValue(tftypes.String, "element2-computed2"),
										},
									),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttribute{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_computed1": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												fwplanmodifier.MatchElementStateForUnknownModifier{
													Expressions: path.Expressions{
														path.MatchRelative().AtParent().AtName("nested_configured1"),
														path.MatchRelative().AtParent().AtName("nested_configured2"),
													},
												},
											},
										},
										"nested_computed2": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
										},
										"nested_configured1": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
										"nested_configured2": testschema.AttributeWithStringPlanModifiers{
											Required: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
							},
						},
					},
				},
			},
			expressions: path.Expressions{
				path.MatchRelative().AtParent().AtName("nested_configured1"),
				path.MatchRelative().AtParent().AtName("nested_configured2"),
			},
			expected: &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: types.StringValue("element2-computed1"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &fwplanmodifier.MatchElementStateForUnknownResponse{
				PlanValue: testCase.request.PlanValue,
			}
			fwplanmodifier.MatchElementStateForUnknownModifier{
				Expressions: testCase.expressions,
			}.PlanModify(context.Background(), testCase.request, resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMatchElementStateForUnknownMissingExpressionsDiag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		expected diag.Diagnostic
	}{
		"test": {
			path: path.Root("test").AtListIndex(0).AtName("nested_test"),
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test").AtListIndex(0).AtName("nested_test"),
				"Invalid Attribute Schema",
				"The MatchElementStateForUnknown() plan modifier has no path expressions. "+
					"At least one path expression must be given for matching the prior state. "+
					"For example:\n\n"+
					"MatchElementStateForUnknown(\n"+
					"  path.MatchRelative().AtParent().AtName(\"another_element_attribute\"),\n"+
					"),\n\n"+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"Path: test[0].nested_test",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwplanmodifier.MatchElementStateForUnknownMissingExpressionsDiag(testCase.path)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMatchElementStateForUnknownOutsideListOrSetDiag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		expected diag.Diagnostic
	}{
		"test": {
			path: path.Root("test").AtMapKey("testkey").AtName("nested_test"),
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test").AtMapKey("testkey").AtName("nested_test"),
				"Invalid Attribute Schema",
				"The MatchElementStateForUnknown() plan modifier is only intended for nested object attributes under a list or set. "+
					"Use the UseStateForUnknown() plan modifier instead. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"Path: test[\"testkey\"].nested_test",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwplanmodifier.MatchElementStateForUnknownOutsideListOrSetDiag(testCase.path)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMatchElementStateForUnknownInvalidExpressionDiag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path       path.Path
		expression path.Expression
		expected   diag.Diagnostic
	}{
		"test": {
			path:       path.Root("test").AtListIndex(0).AtName("nested_test"),
			expression: path.MatchRelative().AtParent().AtParent().AtName("not_test"),
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test").AtListIndex(0).AtName("nested_test"),
				"Invalid Attribute Schema",
				"The MatchElementStateForUnknown() plan modifier was given an invalid path expression. "+
					"Expressions should be relative and match a different, identifying, and configurable attribute within the same nested object. "+
					"For example:\n\n"+
					"MatchElementStateForUnknown(\n"+
					"  path.MatchRelative().AtParent().AtName(\"another_element_attribute\"),\n"+
					"),\n\n"+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"Path: test[0].nested_test\n"+
					"Given Expression: <.<.not_test",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwplanmodifier.MatchElementStateForUnknownInvalidExpressionDiag(testCase.path, testCase.expression)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMatchElementStateForUnknownRootExpressionDiag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path       path.Path
		expression path.Expression
		expected   diag.Diagnostic
	}{
		"test": {
			path:       path.Root("test").AtListIndex(0).AtName("nested_test"),
			expression: path.MatchRoot("test").AtListIndex(0).AtName("not_nested_test"), // not valid for multiple elements
			expected: diag.NewAttributeErrorDiagnostic(
				path.Root("test").AtListIndex(0).AtName("nested_test"),
				"Invalid Attribute Schema",
				"The MatchElementStateForUnknown() plan modifier was given a root path expression. "+
					"Expressions should only be relative and reference attributes at the same level. "+
					"For example:\n\n"+
					"MatchElementStateForUnknown(\n"+
					"  path.MatchRelative().AtParent().AtName(\"another_element_attribute\"),\n"+
					"),\n\n"+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					"Path: test[0].nested_test\n"+
					"Given Expression: test[0].not_nested_test",
			),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwplanmodifier.MatchElementStateForUnknownRootExpressionDiag(testCase.path, testCase.expression)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
