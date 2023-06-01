// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/planmodifierdiag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaModifyPlan(t *testing.T) {
	t.Parallel()

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testCases := map[string]struct {
		req          ModifySchemaPlanRequest
		expectedResp ModifySchemaPlanResponse
	}{
		"config-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"List Type Validation Error",
						"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"expected List value, received tftypes.Value with value: tftypes.String<\"testvalue\">",
					),
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
		},
		"plan-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"List Type Validation Error",
						"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"expected List value, received tftypes.Value with value: tftypes.String<\"testvalue\">",
					),
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
			},
		},
		"state-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"List Type Validation Error",
						"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"expected List value, received tftypes.Value with value: tftypes.String<\"testvalue\">",
					),
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
		},
		"no-plan-modifiers": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.Attribute{
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
		},
		"attribute-plan": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRONE" {
												resp.PlanValue = types.StringValue("TESTATTRTWO")
											}
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRTWO" {
												resp.PlanValue = types.StringValue("MODIFIED_TWO")
											}
										},
									},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRONE" {
												resp.PlanValue = types.StringValue("TESTATTRTWO")
											}
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRTWO" {
												resp.PlanValue = types.StringValue("MODIFIED_TWO")
											}
										},
									},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRONE" {
												resp.PlanValue = types.StringValue("TESTATTRTWO")
											}
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRTWO" {
												resp.PlanValue = types.StringValue("MODIFIED_TWO")
											}
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRONE" {
												resp.PlanValue = types.StringValue("TESTATTRTWO")
											}
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRTWO" {
												resp.PlanValue = types.StringValue("MODIFIED_TWO")
											}
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"attribute-request-private": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									planmodifiers.TestAttrPlanPrivateModifierGet{},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									planmodifiers.TestAttrPlanPrivateModifierGet{},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									planmodifiers.TestAttrPlanPrivateModifierGet{},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									planmodifiers.TestAttrPlanPrivateModifierGet{},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-response-private": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-list-nested-private": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithListPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.List{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithListPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.List{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				State: tfsdk.State{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithListPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.List{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithListPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.List{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-set-nested-private": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithSetPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Set{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithSetPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Set{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				State: tfsdk.State{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithSetPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Set{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithSetPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Set{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-set-nested-usestateforunknown": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_computed": tftypes.String,
											"nested_required": tftypes.String,
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
											"nested_computed": tftypes.String,
											"nested_required": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed": tftypes.String,
												"nested_required": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_computed": tftypes.NewValue(tftypes.String, nil),
											"nested_required": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed": tftypes.String,
												"nested_required": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_computed": tftypes.NewValue(tftypes.String, nil),
											"nested_required": tftypes.NewValue(tftypes.String, "testvalue2"),
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
										"nested_computed": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"nested_required": testschema.Attribute{
											Type:     types.StringType,
											Computed: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
								Required:    true,
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_computed": tftypes.String,
											"nested_required": tftypes.String,
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
											"nested_computed": tftypes.String,
											"nested_required": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed": tftypes.String,
												"nested_required": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_required": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed": tftypes.String,
												"nested_required": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_required": tftypes.NewValue(tftypes.String, "testvalue2"),
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
										"nested_computed": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"nested_required": testschema.Attribute{
											Type:     types.StringType,
											Computed: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
								Required:    true,
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_computed": tftypes.String,
											"nested_required": tftypes.String,
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
											"nested_computed": tftypes.String,
											"nested_required": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed": tftypes.String,
												"nested_required": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_computed": tftypes.NewValue(tftypes.String, "statevalue1"),
											"nested_required": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed": tftypes.String,
												"nested_required": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_computed": tftypes.NewValue(tftypes.String, "statevalue2"),
											"nested_required": tftypes.NewValue(tftypes.String, "testvalue2"),
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
										"nested_computed": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"nested_required": testschema.Attribute{
											Type:     types.StringType,
											Computed: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
								Required:    true,
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Set{
									ElementType: tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"nested_computed": tftypes.String,
											"nested_required": tftypes.String,
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
											"nested_computed": tftypes.String,
											"nested_required": tftypes.String,
										},
									},
								},
								[]tftypes.Value{
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed": tftypes.String,
												"nested_required": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_required": tftypes.NewValue(tftypes.String, "testvalue1"),
										},
									),
									tftypes.NewValue(
										tftypes.Object{
											AttributeTypes: map[string]tftypes.Type{
												"nested_computed": tftypes.String,
												"nested_required": tftypes.String,
											},
										},
										map[string]tftypes.Value{
											"nested_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
											"nested_required": tftypes.NewValue(tftypes.String, "testvalue2"),
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
										"nested_computed": testschema.AttributeWithStringPlanModifiers{
											Computed: true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"nested_required": testschema.Attribute{
											Type:     types.StringType,
											Computed: true,
										},
									},
								},
								NestingMode: fwschema.NestingModeSet,
								Required:    true,
							},
						},
					},
				},
				Diagnostics: diag.Diagnostics{
					planmodifierdiag.UseStateForUnknownUnderListOrSet(
						path.Root("test").AtSetValue(
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_computed": types.StringType,
									"nested_required": types.StringType,
								},
								map[string]attr.Value{
									"nested_computed": types.StringUnknown(),
									"nested_required": types.StringValue("testvalue1"),
								},
							),
						).AtName("nested_computed"),
					),
				},
			},
		},
		"attribute-map-nested-private": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithMapPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Map{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithMapPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Map{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				State: tfsdk.State{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithMapPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Map{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
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
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithMapPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"nested_attr": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Map{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-single-nested-private": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testing": tftypes.String,
									},
								},
							},
						}, map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testing": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"testing": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithObjectPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"testing": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Object{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testing": tftypes.String,
									},
								},
							},
						}, map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testing": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"testing": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithObjectPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"testing": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Object{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testing": tftypes.String,
									},
								},
							},
						}, map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testing": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"testing": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithObjectPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"testing": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Object{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: privatestate.EmptyProviderData(context.Background()),
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test": tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testing": tftypes.String,
									},
								},
							},
						}, map[string]tftypes.Value{
							"test": tftypes.NewValue(
								tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"testing": tftypes.String,
									},
								},
								map[string]tftypes.Value{
									"testing": tftypes.NewValue(tftypes.String, "testvalue"),
								},
							),
						},
					),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.NestedAttributeWithObjectPlanModifiers{
								NestedObject: testschema.NestedAttributeObject{
									Attributes: map[string]fwschema.Attribute{
										"testing": testschema.AttributeWithStringPlanModifiers{
											Required: true,
											PlanModifiers: []planmodifier.String{
												planmodifiers.TestAttrPlanPrivateModifierGet{},
											},
										},
									},
								},
								Required: true,
								PlanModifiers: []planmodifier.Object{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"requires-replacement": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "newtestvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "newtestvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "newtestvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"requires-replacement-passthrough": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRONE" {
												resp.PlanValue = types.StringValue("TESTATTRTWO")
											}
										},
									},
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRONE" {
												resp.PlanValue = types.StringValue("TESTATTRTWO")
											}
										},
									},
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRONE" {
												resp.PlanValue = types.StringValue("TESTATTRTWO")
											}
										},
									},
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRTWO"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											if req.PlanValue.ValueString() == "TESTATTRONE" {
												resp.PlanValue = types.StringValue("TESTATTRTWO")
											}
										},
									},
								},
							},
						},
					},
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
			},
		},
		"requires-replacement-unset": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.RequiresReplace = false
										},
									},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.RequiresReplace = false
										},
									},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.RequiresReplace = false
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.RequiresReplace = false
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"warnings": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
										},
									},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
										},
									},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					// Diagnostics.Append() deduplicates, so the warning will only
					// be here once unless the test implementation is changed to
					// different modifiers or the modifier itself is changed.
					diag.NewWarningDiagnostic(
						"Warning diag",
						"This is a warning",
					),
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddWarning("Warning diag", "This is a warning")
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddError("Error diag", "This is an error")
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddError("Error diag", "This is an error")
										},
									},
								},
							},
						},
					},
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddError("Error diag", "This is an error")
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddError("Error diag", "This is an error")
										},
									},
								},
							},
						},
					},
				},
				State: tfsdk.State{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddError("Error diag", "This is an error")
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddError("Error diag", "This is an error")
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: testschema.Schema{
						Attributes: map[string]fwschema.Attribute{
							"test": testschema.AttributeWithStringPlanModifiers{
								Required: true,
								PlanModifiers: []planmodifier.String{
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddError("Error diag", "This is an error")
										},
									},
									testplanmodifier.String{
										PlanModifyStringMethod: func(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
											resp.Diagnostics.AddError("Error diag", "This is an error")
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			got := ModifySchemaPlanResponse{
				Plan:    tc.req.Plan,
				Private: tc.req.Private,
			}

			SchemaModifyPlan(ctx, tc.req.Plan.Schema, tc.req, &got)

			if diff := cmp.Diff(tc.expectedResp, got, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("Unexpected response (-wanted, +got): %s", diff)
			}
		})
	}
}
