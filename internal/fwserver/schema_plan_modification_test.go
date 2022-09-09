package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaModifyPlan(t *testing.T) {
	t.Parallel()

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
						"Configuration Read Error",
						"An unexpected error was encountered trying to convert an attribute value from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"Error: can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
		},
		"config-error-previous-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Configuration Read Error",
						"An unexpected error was encountered trying to convert an attribute value from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"Error: can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
						"Plan Read Error",
						"An unexpected error was encountered trying to convert an attribute value from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"Error: can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
			},
		},
		"plan-error-previous-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"Plan Read Error",
						"An unexpected error was encountered trying to convert an attribute value from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"Error: can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
						"State Read Error",
						"An unexpected error was encountered trying to convert an attribute value from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"Error: can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
			},
		},
		"state-error-previous-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "testvalue"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.ListType{ElemType: types.StringType},
								Required: true,
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
					diag.NewAttributeErrorDiagnostic(
						path.Root("test"),
						"State Read Error",
						"An unexpected error was encountered trying to convert an attribute value from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
							"Error: can't use tftypes.String<\"testvalue\"> as value of List with ElementType types.primitive, can only use tftypes.String values",
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									planmodifiers.TestAttrPlanValueModifierTwo{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									planmodifiers.TestAttrPlanValueModifierTwo{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									planmodifiers.TestAttrPlanValueModifierTwo{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									planmodifiers.TestAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
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
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"nested_computed": {
										Type:     types.StringType,
										Computed: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											resource.UseStateForUnknown(),
										},
									},
									"nested_required": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"nested_computed": {
										Type:     types.StringType,
										Computed: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											resource.UseStateForUnknown(),
										},
									},
									"nested_required": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"nested_computed": {
										Type:     types.StringType,
										Computed: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											resource.UseStateForUnknown(),
										},
									},
									"nested_required": {
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
									"nested_computed": {
										Type:     types.StringType,
										Computed: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											resource.UseStateForUnknown(),
										},
									},
									"nested_required": {
										Type:     types.StringType,
										Required: true,
									},
								}),
								Required: true,
							},
						},
					},
				},
				Private: testEmptyProviderData,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
									"nested_attr": {
										Type:     types.StringType,
										Required: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: tfsdk.AttributePlanModifiers{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"testing": {
										Type:     types.StringType,
										Optional: true,
										PlanModifiers: tfsdk.AttributePlanModifiers{
											planmodifiers.TestAttrPlanPrivateModifierGet{},
										},
									},
								}),
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanPrivateModifierSet{},
								},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-plan-previous-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTATTRONE"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									planmodifiers.TestAttrPlanValueModifierTwo{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									planmodifiers.TestAttrPlanValueModifierTwo{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									planmodifiers.TestAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									planmodifiers.TestAttrPlanValueModifierTwo{},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"requires-replacement-previous-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "newtestvalue"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Previous error diag",
						"This was a previous error",
					),
				},
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "newtestvalue"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestAttrPlanValueModifierOne{},
									resource.RequiresReplace(),
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
									planmodifiers.TestAttrPlanValueModifierOne{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
									planmodifiers.TestAttrPlanValueModifierOne{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
									planmodifiers.TestRequiresReplaceFalseModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
									planmodifiers.TestRequiresReplaceFalseModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.RequiresReplace(),
									planmodifiers.TestRequiresReplaceFalseModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
				},
				Private: testEmptyProviderData,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestWarningDiagModifier{},
									planmodifiers.TestWarningDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestWarningDiagModifier{},
									planmodifiers.TestWarningDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestWarningDiagModifier{},
									planmodifiers.TestWarningDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestWarningDiagModifier{},
									planmodifiers.TestWarningDiagModifier{},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"warnings-previous-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestWarningDiagModifier{},
									planmodifiers.TestWarningDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestWarningDiagModifier{},
									planmodifiers.TestWarningDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestWarningDiagModifier{},
									planmodifiers.TestWarningDiagModifier{},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
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
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestWarningDiagModifier{},
									planmodifiers.TestWarningDiagModifier{},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestErrorDiagModifier{},
									planmodifiers.TestErrorDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestErrorDiagModifier{},
									planmodifiers.TestErrorDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestErrorDiagModifier{},
									planmodifiers.TestErrorDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestErrorDiagModifier{},
									planmodifiers.TestErrorDiagModifier{},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"error-previous-error": {
			req: ModifySchemaPlanRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestErrorDiagModifier{},
									planmodifiers.TestErrorDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestErrorDiagModifier{},
									planmodifiers.TestErrorDiagModifier{},
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
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestErrorDiagModifier{},
									planmodifiers.TestErrorDiagModifier{},
								},
							},
						},
					},
				},
			},
			expectedResp: ModifySchemaPlanResponse{
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
				Plan: tfsdk.Plan{
					Raw: tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"test": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"test": tftypes.NewValue(tftypes.String, "TESTDIAG"),
					}),
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Type:     types.StringType,
								Required: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									planmodifiers.TestErrorDiagModifier{},
									planmodifiers.TestErrorDiagModifier{},
								},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		if name != "attribute-set-nested-usestateforunknown" {
			continue
		}

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
