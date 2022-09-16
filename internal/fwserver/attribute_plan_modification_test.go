package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAttributeModifyPlan(t *testing.T) {
	t.Parallel()

	testProviderKeyValue := privatestate.MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testProviderData := privatestate.MustProviderData(context.Background(), testProviderKeyValue)

	testEmptyProviderData := privatestate.EmptyProviderData(context.Background())

	testCases := map[string]struct {
		attribute    fwschema.Attribute
		req          tfsdk.ModifyAttributePlanRequest
		expectedResp ModifyAttributePlanResponse
	}{
		"no-plan-modifiers": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "testvalue"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "testvalue"},
				AttributeState:  types.String{Value: "testvalue"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
			},
		},
		"attribute-plan": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanValueModifierOne{},
					planmodifiers.TestAttrPlanValueModifierTwo{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "TESTATTRONE"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "TESTATTRONE"},
				AttributeState:  types.String{Value: "TESTATTRONE"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "MODIFIED_TWO"},
				Private:       testEmptyProviderData,
			},
		},
		"attribute-request-private": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanPrivateModifierGet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "TESTATTRONE"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "TESTATTRONE"},
				AttributeState:  types.String{Value: "TESTATTRONE"},
				Private:         testProviderData,
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTATTRONE"},
				Private:       testProviderData,
			},
		},
		"attribute-response-private": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanPrivateModifierSet{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "TESTATTRONE"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "TESTATTRONE"},
				AttributeState:  types.String{Value: "TESTATTRONE"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTATTRONE"},
				Private:       testProviderData,
			},
		},
		"attribute-list-nested-private": {
			attribute: tfsdk.Attribute{
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
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributeState: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.List{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-set-nested-private": {
			attribute: tfsdk.Attribute{
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
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributeState: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-set-nested-usestateforunknown": {
			attribute: tfsdk.Attribute{
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
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Null: true},
								"nested_required": types.String{Value: "testvalue1"},
							},
						},
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Null: true},
								"nested_required": types.String{Value: "testvalue2"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Unknown: true},
								"nested_required": types.String{Value: "testvalue1"},
							},
						},
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Unknown: true},
								"nested_required": types.String{Value: "testvalue2"},
							},
						},
					},
				},
				AttributeState: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Value: "statevalue1"},
								"nested_required": types.String{Value: "testvalue1"},
							},
						},
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Value: "statevalue2"},
								"nested_required": types.String{Value: "testvalue2"},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Set{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_computed": types.StringType,
							"nested_required": types.StringType,
						},
					},
					Elems: []attr.Value{
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Value: "statevalue1"},
								"nested_required": types.String{Value: "testvalue1"},
							},
						},
						types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_computed": types.StringType,
								"nested_required": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_computed": types.String{Value: "statevalue2"},
								"nested_required": types.String{Value: "testvalue2"},
							},
						},
					},
				},
				Private: testEmptyProviderData,
			},
		},
		"attribute-map-nested-private": {
			attribute: tfsdk.Attribute{
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
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Map{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: map[string]attr.Value{
						"testkey": types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Map{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: map[string]attr.Value{
						"testkey": types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				AttributeState: types.Map{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: map[string]attr.Value{
						"testkey": types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Map{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"nested_attr": types.StringType,
						},
					},
					Elems: map[string]attr.Value{
						"testkey": types.Object{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
							Attrs: map[string]attr.Value{
								"nested_attr": types.String{Value: "testvalue"},
							},
						},
					},
				},
				Private: testProviderData,
			},
		},
		"attribute-single-nested-private": {
			attribute: tfsdk.Attribute{
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
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.Object{
					AttrTypes: map[string]attr.Type{
						"testing": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"testing": types.String{Value: "testvalue"},
					},
				},
				AttributePath: path.Root("test"),
				AttributePlan: types.Object{
					AttrTypes: map[string]attr.Type{
						"testing": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"testing": types.String{Value: "testvalue"},
					},
				},
				AttributeState: types.Object{
					AttrTypes: map[string]attr.Type{
						"testing": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"testing": types.String{Value: "testvalue"},
					},
				},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.Object{
					AttrTypes: map[string]attr.Type{
						"testing": types.StringType,
					},
					Attrs: map[string]attr.Value{
						"testing": types.String{Value: "testvalue"},
					},
				},
				Private: testProviderData,
			},
		},
		"requires-replacement": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "newtestvalue"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "newtestvalue"},
				AttributeState:  types.String{Value: "testvalue"},
				// resource.RequiresReplace() requires non-null plan
				// and state.
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
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "newtestvalue"},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"requires-replacement-passthrough": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestAttrPlanValueModifierOne{},
					resource.RequiresReplace(),
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "TESTATTRONE"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "TESTATTRONE"},
				AttributeState:  types.String{Value: "TESTATTRONE"},
				// resource.RequiresReplace() requires non-null plan
				// and state.
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
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTATTRTWO"},
				RequiresReplace: path.Paths{
					path.Root("test"),
				},
				Private: testEmptyProviderData,
			},
		},
		"requires-replacement-unset": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.RequiresReplace(),
					planmodifiers.TestRequiresReplaceFalseModifier{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "testvalue"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "testvalue"},
				AttributeState:  types.String{Value: "testvalue"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "testvalue"},
				Private:       testEmptyProviderData,
			},
		},
		"warnings": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestWarningDiagModifier{},
					planmodifiers.TestWarningDiagModifier{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "TESTDIAG"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "TESTDIAG"},
				AttributeState:  types.String{Value: "TESTDIAG"},
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
				Private: testEmptyProviderData,
			},
		},
		"error": {
			attribute: tfsdk.Attribute{
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					planmodifiers.TestErrorDiagModifier{},
					planmodifiers.TestErrorDiagModifier{},
				},
			},
			req: tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: types.String{Value: "TESTDIAG"},
				AttributePath:   path.Root("test"),
				AttributePlan:   types.String{Value: "TESTDIAG"},
				AttributeState:  types.String{Value: "TESTDIAG"},
			},
			expectedResp: ModifyAttributePlanResponse{
				AttributePlan: types.String{Value: "TESTDIAG"},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Error diag",
						"This is an error",
					),
				},
				Private: testEmptyProviderData,
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			got := ModifyAttributePlanResponse{
				AttributePlan: tc.req.AttributePlan,
				Private:       tc.req.Private,
			}

			AttributeModifyPlan(ctx, tc.attribute, tc.req, &got)

			if diff := cmp.Diff(tc.expectedResp, got, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("Unexpected response (-wanted, +got): %s", diff)
			}
		})
	}
}
